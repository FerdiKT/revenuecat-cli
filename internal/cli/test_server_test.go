package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func newRevenueCatFixtureServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/proj_123/apps":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{
				"id":   "app_1",
				"name": "iOS App",
				"app_store": map[string]any{
					"bundle_id": "app.ferdi.headson",
				},
			}}))
		case "/v2/projects/proj_123/entitlements":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "ent_1", "lookup_key": "pro"}}))
		case "/v2/projects/proj_123/products":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "prod_1", "store_identifier": "monthly"}}))
		case "/v2/projects/proj_123/offerings":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "off_1", "lookup_key": "default"}}))
		case "/v2/projects/proj_123/offerings/off_1/packages":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "pkg_1", "lookup_key": "$rc_monthly"}}))
		case "/v2/projects/proj_123/metrics/overview":
			_ = json.NewEncoder(w).Encode(map[string]any{"revenue": 1000, "trials": 12})
		case "/v2/projects/proj_123/charts/trials":
			_ = json.NewEncoder(w).Encode(map[string]any{"chart_name": "trials", "values": []any{map[string]any{"value": 12}}})
		case "/v2/projects/proj_123/charts/revenue":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chart_name": "revenue",
				"segments": []any{
					map[string]any{"id": "US", "display_name": "United States"},
					map[string]any{"id": "TR", "display_name": "Turkey"},
					map[string]any{"id": "OTHER", "display_name": "Other"},
				},
				"measures": []any{
					map[string]any{"id": "revenue"},
					map[string]any{"id": "transactions"},
				},
				"summary": map[string]any{
					"US":    map[string]any{"revenue": 1234.5, "transactions": 18},
					"TR":    map[string]any{"revenue": 87, "transactions": 4},
					"OTHER": map[string]any{"revenue": 11, "transactions": 1},
				},
			})
		case "/v2/projects/proj_123/customers":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "cust_1", "app_user_id": "user-1"}}))
		case "/v2/projects/proj_123/customers/cust_1/subscriptions":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "sub_1", "status": "active"}}))
		case "/v2/projects/proj_123/customers/cust_1/purchases":
			_ = json.NewEncoder(w).Encode(listResponse([]map[string]any{{"id": "pur_1", "product_id": "prod_1"}}))
		default:
			http.NotFound(w, r)
		}
	}))
}

func listResponse(items []map[string]any) map[string]any {
	anyItems := make([]any, 0, len(items))
	for _, item := range items {
		anyItems = append(anyItems, item)
	}
	return map[string]any{
		"object": "list",
		"items":  anyItems,
	}
}
