package internal

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stackvista/sts-opentelemetry-collector/connector/topologyconnector/metrics"
	"github.com/stackvista/sts-opentelemetry-collector/extension/settingsproviderextension/generated/settingsproto"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func readStackPackMapping(t *testing.T, name string) settingsproto.OtelComponentMapping {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(os.Getenv("RESOURCE_COLLECTION_STACKPACK"),
		"settings", "component-mappings", name+".sty"))
	require.NoError(t, err)
	var document struct {
		Nodes []map[string]any `yaml:"nodes"`
	}
	require.NoError(t, yaml.Unmarshal(body, &document))
	node := document.Nodes[0]
	expression := func(value any) map[string]any {
		return map[string]any{"expression": value}
	}
	for _, variable := range node["vars"].([]any) {
		entry := variable.(map[string]any)
		entry["value"] = expression(entry["value"])
	}
	log := node["input"].(map[string]any)["resource"].(map[string]any)["scope"].(map[string]any)["log"].(map[string]any)
	for _, key := range []string{"condition", "action"} {
		log[key] = expression(log[key])
	}
	output := node["output"].(map[string]any)
	for _, key := range []string{"identifier", "name", "typeName"} {
		output[key] = expression(output[key])
	}
	for _, field := range []string{"required", "optional"} {
		block := output[field].(map[string]any)
		for _, key := range []string{"configuration", "status"} {
			if value, found := block[key]; found {
				block[key] = expression(value)
			}
		}
		for _, tag := range block["tags"].([]any) {
			entry := tag.(map[string]any)
			entry["source"] = expression(entry["source"])
		}
	}
	body, err = json.Marshal(node)
	require.NoError(t, err)
	var mapping settingsproto.OtelComponentMapping
	require.NoError(t, json.Unmarshal(body, &mapping))
	return mapping
}

func TestStackPackResourceCollections(t *testing.T) {
	evaluator, err := NewCELEvaluator(context.Background(), metrics.MeteredCacheSettings{
		Name: "stackpack_resource_collections", EnableMetrics: false,
	})
	require.NoError(t, err)
	mapper := NewMapper(context.Background(),
		metrics.MeteredCacheSettings{Name: "tags", EnableMetrics: false},
		metrics.MeteredCacheSettings{Name: "templates", EnableMetrics: false})
	specialized := readStackPackMapping(t, "resource-collections")
	generic := readStackPackMapping(t, "custom-resources")
	tests := []struct {
		name, ready, reason, expected string
		observed, conditionGeneration int64
	}{
		{"no status", "", "", "Awaiting reconciliation", -1, -1},
		{"empty status", "", "", "Awaiting reconciliation", 0, 0},
		{"missing Ready", "", "", "Awaiting reconciliation", 2, 2},
		{"collecting", "True", "CollectionActive", "Collecting", 2, 2},
		{"pending API", "False", "ResourcesPending", "Pending", 2, 2},
		{"blocked reads", "False", "CollectionBlocked", "Blocked", 2, 2},
		{"blocked policy", "False", "InvalidSelection", "Blocked", 2, 2},
		{"unknown readiness", "Unknown", "Unknown", "Unknown", 2, 2},
		{"stale status", "True", "CollectionActive", "Awaiting reconciliation", 1, 1},
		{"stale condition", "True", "CollectionActive", "Awaiting reconciliation", 2, 1},
		{"missing condition generation", "True", "CollectionActive", "Awaiting reconciliation", 2, -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := map[string]any{
				"apiVersion": "observability.suse.com/v1alpha1", "kind": "ResourceCollection",
				"metadata": map[string]any{"name": "virtualization", "generation": int64(2)},
				"spec": map[string]any{"resources": []any{
					map[string]any{"group": "kubevirt.io", "resource": "virtualmachines"},
				}},
			}
			if test.observed >= 0 {
				status := map[string]any{}
				if test.observed > 0 {
					status["observedGeneration"] = test.observed
				}
				if test.ready != "" {
					condition := map[string]any{
						"type": "Ready", "status": test.ready, "reason": test.reason, "message": "Fixture feedback.",
					}
					if test.conditionGeneration >= 0 {
						condition["observedGeneration"] = test.conditionGeneration
					}
					status["conditions"] = []any{
						map[string]any{"type": "Other", "status": "False"},
						condition,
					}
				}
				object["status"] = status
			}
			evalContext := NewLogEvalContext(NewLog("KubernetesCustomResourceEvent",
				map[string]any{"type": "MODIFIED", "object": object},
				map[string]any{"k8s.resource.group": "observability.suse.com",
					"k8s.resource.kind": "ResourceCollection", "k8s.object.name": "virtualization"}),
				nil, NewResource(map[string]any{"k8s.cluster.name": "fixture"}))
			selected, err := evaluator.EvalBooleanExpression(*specialized.Input.Resource.Scope.Log.Condition, evalContext)
			require.NoError(t, err)
			require.True(t, selected)
			selected, err = evaluator.EvalBooleanExpression(*generic.Input.Resource.Scope.Log.Condition, evalContext)
			require.NoError(t, err)
			require.False(t, selected)
			variables, errors := EvalVariables(evaluator, evalContext, specialized.Vars)
			require.Empty(t, errors)
			component, errors := mapper.MapComponent(&specialized, evaluator, evalContext.CloneWithVariables(variables))
			require.Empty(t, errors)
			require.Equal(t, "urn:k8s:cr:/cluster/fixture:namespace/:group/observability.suse.com:kind/ResourceCollection:name/virtualization", component.ExternalId)
			require.Equal(t, "custom resource", component.TypeName)
			require.Contains(t, component.Tags, "k8s.collection.state:"+test.expected)
			require.Contains(t, component.Tags, "k8s.collection.selections:1")
			if test.expected == "Awaiting reconciliation" {
				require.Contains(t, component.Tags, "k8s.collection.reason:")
				if test.reason != "" {
					require.NotContains(t, component.Tags, "k8s.collection.reason:"+test.reason)
				}
				require.Contains(t, component.Tags, "k8s.collection.message:Waiting for feedback on the current declaration.")
			} else {
				require.Contains(t, component.Tags, "k8s.collection.reason:"+test.reason)
			}
			require.Equal(t, object["spec"], component.ResourceDefinition.AsMap()["spec"])
			require.NotContains(t, component.ResourceDefinition.AsMap(), "metadata")
			require.NotContains(t, component.ResourceDefinition.AsMap(), "status")
			if status, ok := object["status"]; ok {
				expected, err := json.Marshal(status)
				require.NoError(t, err)
				actual, err := json.Marshal(component.StatusData.AsMap())
				require.NoError(t, err)
				require.JSONEq(t, string(expected), string(actual))
			}
		})
	}
	t.Run("ordinary CRs retain their generic mapping", func(t *testing.T) {
		evalContext := NewLogEvalContext(NewLog("KubernetesCustomResourceEvent",
			map[string]any{"type": "ADDED"}, map[string]any{
				"k8s.resource.group": "kubevirt.io", "k8s.resource.kind": "VirtualMachine",
			}), nil, NewResource(map[string]any{}))
		for _, mapping := range []settingsproto.OtelComponentMapping{generic, specialized} {
			selected, err := evaluator.EvalBooleanExpression(*mapping.Input.Resource.Scope.Log.Condition, evalContext)
			require.NoError(t, err)
			require.Equal(t, mapping.Identifier == generic.Identifier, selected)
		}
	})
	t.Run("delete requires identity only", func(t *testing.T) {
		evalContext := NewLogEvalContext(NewLog("KubernetesCustomResourceEvent",
			map[string]any{"type": "DELETED"}, map[string]any{
				"k8s.resource.group": "observability.suse.com", "k8s.resource.kind": "ResourceCollection",
				"k8s.object.name": "virtualization",
			}), nil, NewResource(map[string]any{"k8s.cluster.name": "fixture"}))
		action, err := evaluator.EvalStringExpression(*specialized.Input.Resource.Scope.Log.Action, evalContext)
		require.NoError(t, err)
		require.Equal(t, "DELETE", action)
		needed := CollectVarReferences(evaluator, specialized.Output.Identifier)
		variables := FilterVarsByName(specialized.Vars, needed)
		resolved, errors := EvalVariables(evaluator, evalContext, &variables)
		require.Empty(t, errors)
		identifier, err := evaluator.EvalStringExpression(specialized.Output.Identifier, evalContext.CloneWithVariables(resolved))
		require.NoError(t, err)
		require.True(t, strings.HasSuffix(identifier, ":name/virtualization"))
	})
}
