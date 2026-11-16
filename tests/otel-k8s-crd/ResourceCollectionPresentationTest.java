import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.stackstate.api.CnfSubQueryMatcher;
import com.stackstate.api.generated.model.ComponentField;
import com.stackstate.api.generated.model.ErrorField;
import com.stackstate.api.generated.model.TextField;
import com.stackstate.api.service.components.FieldContext;
import com.stackstate.api.service.components.FieldExtractor;
import com.stackstate.api.service.components.ParsedComponentField;
import com.stackstate.api.service.components.RelatedResourceExtractor;
import com.stackstate.api.service.overview.OverviewColumnDefinitionWithExpressions;
import com.stackstate.api.service.overview.OverviewPresentationDef;
import com.stackstate.api.service.overview.OverviewPresentationDefs;
import com.stackstate.api.service.overview.PresentationOverviewBuilder;
import com.stackstate.domain.state.HealthStateValue;
import com.stackstate.dto.dtos.componentpresentation.ComponentPresentation;
import com.stackstate.metrics.promql.PromQLParser;
import com.stackstate.metrics.promql.PromQLParserImpl;
import com.stackstate.settings.protocol.ComponentPresentationQueryBinding;
import com.stackstate.settings.protocol.ComponentPresentationRank;
import com.stackstate.settings.protocol.PresentationDefinition;
import com.stackstate.settings.provider.SettingsProvider;
import com.stackstate.stql.ErrorMessageParser$;
import org.apache.pekko.event.LoggingAdapter;
import org.apache.pekko.event.NoLogging$;
import scala.Option;
import scala.Tuple2;
import scala.collection.immutable.List;
import scala.collection.immutable.Seq;
import scala.collection.immutable.Set;
import scala.jdk.javaapi.CollectionConverters;
import spray.json.JsonParser;
import spray.json.ParserInput;

import java.lang.reflect.Method;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Map;

public class ResourceCollectionPresentationTest implements PresentationOverviewBuilder {
    public SettingsProvider settingsProvider() { return null; }
    public PromQLParser promqlParser() { return new PromQLParserImpl(); }
    public LoggingAdapter log() { return NoLogging$.MODULE$; }

    private static void check(boolean value, String message) {
        if (!value) throw new AssertionError(message);
    }

    private static <T> Seq<T> seq(java.util.List<T> values) {
        return List.from(CollectionConverters.asScala(values));
    }

    private static ComponentPresentation read(Path path) throws Exception {
        JsonNode node = new ObjectMapper(new YAMLFactory()).readTree(path.toFile()).get("nodes").get(0);
        PresentationDefinition definition = PresentationDefinition.formatter().read(
            JsonParser.apply(ParserInput.apply(node.get("presentation").toString())));
        return new ComponentPresentation(
            Option.empty(), Option.empty(), node.get("name").asText(), Option.empty(),
            Option.apply(node.get("identifier").asText()), Option.empty(),
            new ComponentPresentationQueryBinding(node.get("binding").get("query").asText()),
            Option.apply(new ComponentPresentationRank(node.get("rank").get("specificity").asDouble())),
            Option.empty(), definition);
    }

    private static boolean matches(ComponentPresentation candidate, ComponentPresentation base) {
        return CnfSubQueryMatcher.isSubquery(
            ErrorMessageParser$.MODULE$.parseTry(candidate.binding().query()).get(),
            ErrorMessageParser$.MODULE$.parseTry(base.binding().query()).get());
    }

    private static FieldContext context(String state, boolean observed) {
        var tags = new ArrayList<String>(java.util.List.of(
            "stackpack:otel-k8s-crd", "source-type:custom resource",
            "k8s.resource.group:observability.suse.com", "k8s.resource.kind:ResourceCollection",
            "cluster-name:fixture", "k8s.creationTimestamp:2026-11-13T12:00:00Z",
            "k8s.collection.state:" + state, "k8s.collection.reason:FixtureReason",
            "k8s.collection.message:Fixture feedback.", "k8s.collection.generation:2",
            "k8s.collection.selections:1"));
        if (observed) tags.add("k8s.collection.observed-generation:2");
        return new FieldContext(
            1, "virtualization",
            Set.from(CollectionConverters.asScala(java.util.Set.of(
                "urn:k8s:cr:/cluster/fixture:namespace/:group/observability.suse.com:kind/ResourceCollection:name/virtualization"))),
            Set.from(CollectionConverters.asScala(tags)),
            scala.collection.immutable.Map.from(CollectionConverters.asScala(Map.<String, String>of())),
            "", "", "custom resource", HealthStateValue.CLEAR, Option.empty());
    }

    public static void main(String[] args) throws Exception {
        Path presentations = Path.of(args[0], "settings", "presentations");
        ComponentPresentation specialized = read(presentations.resolve("resource-collections.sty"));
        ComponentPresentation generic = read(presentations.resolve("custom-resources.sty"));
        ComponentPresentation filters = read(presentations.resolve("common-filters.sty"));
        check(!matches(generic, specialized), "Generic CR presentation also matches ResourceCollections");
        check(matches(filters, specialized), "Common cluster filter does not match");
        var matched = seq(java.util.List.of(specialized, filters));
        var parsed = FieldExtractor.parseFields(matched, Option.empty());
        check(parsed._1().isEmpty(), "Highlight parse errors: " + parsed._1());
        var fieldIds = new ArrayList<String>();
        for (ParsedComponentField field : CollectionConverters.asJava(parsed._2())) {
            fieldIds.add(field.fieldId());
        }
        check(!fieldIds.contains("namespace"), "Cluster-scoped declarations show namespace");
        check(fieldIds.containsAll(java.util.List.of(
            "collection_state", "collection_reason", "collection_message", "collection_generation",
            "collection_observed_generation", "collection_selections", "kind", "cluster", "age", "labels")),
            "Missing highlight fields: " + fieldIds);
        System.out.println("PASS presentation matching and highlight composition");

        for (String state : java.util.List.of("Collecting", "Pending", "Blocked", "Unknown", "Awaiting reconciliation")) {
            var evaluated = FieldExtractor.evaluate(parsed._2(), context(state, !state.equals("Awaiting reconciliation")),
                new PromQLParserImpl());
            var values = new java.util.HashMap<String, String>();
            for (ComponentField field : CollectionConverters.asJava(evaluated)) {
                check(!(field instanceof ErrorField), "Highlight evaluation error for " + state + ": " + field);
                if (field instanceof TextField text) values.put(text.fieldId(), text.value());
            }
            check(state.equals(values.get("collection_state")), "Incorrect collection state: " + values);
            check("fixture".equals(values.get("cluster")), "Missing cluster: " + values);
            check("Fixture feedback.".equals(values.get("collection_message")), "Missing feedback: " + values);
            System.out.println("PASS highlight values: " + state);
        }

        Method merge = PresentationOverviewBuilder.class.getDeclaredMethod(
            "mergePresentationsAndBuildColumnDef", OverviewPresentationDefs.class);
        merge.setAccessible(true);
        var base = new OverviewPresentationDef(specialized.identifier().get(), specialized.rank().get().specificity(),
            ErrorMessageParser$.MODULE$.parseTry(specialized.binding().query()).get(),
            specialized.presentation().overview().get());
        @SuppressWarnings("unchecked")
        var overview = (Tuple2<Seq<?>, Seq<OverviewColumnDefinitionWithExpressions>>) merge.invoke(
            new ResourceCollectionPresentationTest(), new OverviewPresentationDefs(base, seq(java.util.List.of())));
        check(overview._1().isEmpty(), "Overview projection errors: " + overview._1());
        var columns = new ArrayList<String>();
        for (var column : CollectionConverters.asJava(overview._2())) columns.add(column.columnId());
        check(columns.equals(java.util.List.of(
            "cr_name", "collection_state", "collection_reason", "collection_selections", "cr_cluster", "cr_age")),
            "Incorrect overview columns: " + columns);
        System.out.println("PASS overview column composition");

        var relations = RelatedResourceExtractor.resolve(matched, Option.empty(), context("Collecting", true));
        check(relations.size() == 1, "Missing or duplicate CRD detail link: " + relations);
        check(relations.head().toString().contains(
            "urn:k8s:cr:/cluster/fixture:namespace/:group/observability.suse.com:kind/ResourceCollection:name/virtualization"),
            "CRD query did not specialize the component identifier: " + relations);
        System.out.println("PASS CRD related-resource query specialization");
    }
}
