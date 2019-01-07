package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/inmem"
	"github.com/influxdata/platform/mock"
	platformtesting "github.com/influxdata/platform/testing"
	"github.com/julienschmidt/httprouter"
)

const (
	targetOneIDString = "0000000000000111"
	// targetTwoIDString = "0000000000000222"
)

var (
	targetOneID = platformtesting.MustIDBase16(targetOneIDString)
	// targetTwoID = platformtesting.MustIDBase16(targetTwoIDString)
)

func TestService_handleGetScraperTargets(t *testing.T) {}

func TestService_handleGetScraperTarget(t *testing.T) {
	type fields struct {
		Service platform.ScraperTargetStoreService
	}

	type args struct {
		id string
	}

	type wants struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "get a scraper target by id",
			fields: fields{
				Service: &mock.ScraperTargetStoreService{
					GetTargetByIDF: func(ctx context.Context, id platform.ID) (*platform.ScraperTarget, error) {
						if id == targetOneID {
							return &platform.ScraperTarget{
								ID:         targetOneID,
								Name:       "target-1",
								Type:       platform.PrometheusScraperType,
								URL:        "www.some.url",
								OrgName:    "org-name",
								BucketName: "bkt-name",
							}, nil
						}
						return nil, fmt.Errorf("not found")
					},
				},
			},
			args: args{
				id: targetOneIDString,
			},
			wants: wants{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"id": "%[1]s", "name": "target-1", "type": "prometheus", "url": "www.some.url", "bucket": "bkt-name", "org": "org-name", "links": {"self": "/api/v2/scrapertargets/%[1]s"}}`, targetOneIDString),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewScraperHandler()
			h.ScraperStorageService = tt.fields.Service

			r := httptest.NewRequest("GET", "http://any.tld", nil)

			r = r.WithContext(context.WithValue(
				context.Background(),
				httprouter.ParamsKey,
				httprouter.Params{
					{
						Key:   "id",
						Value: tt.args.id,
					},
				}))

			w := httptest.NewRecorder()

			h.handleGetScraperTarget(w, r)

			res := w.Result()
			content := res.Header.Get("Content-Type")
			body, _ := ioutil.ReadAll(res.Body)

			if res.StatusCode != tt.wants.statusCode {
				t.Errorf("%q. handleGetScraperTarget() = %v, want %v", tt.name, res.StatusCode, tt.wants.statusCode)
			}
			if tt.wants.contentType != "" && content != tt.wants.contentType {
				t.Errorf("%q. handleGetScraperTarget() = %v, want %v", tt.name, content, tt.wants.contentType)
			}
			if eq, diff, _ := jsonEqual(string(body), tt.wants.body); tt.wants.body != "" && !eq {
				t.Errorf("%q. handleGetScraperTarget() = ***%s***", tt.name, diff)
			}
		})
	}
}

func TestService_handlePatchScraperTarget(t *testing.T) {

}

func TestService_handleDeleteScraperTarget(t *testing.T) {

}

func TestService_handlePostScraperTarget(t *testing.T) {

}

func initScraperService(f platformtesting.TargetFields, t *testing.T) (platform.ScraperTargetStoreService, string, func()) {
	t.Helper()
	svc := inmem.NewService()
	svc.IDGenerator = f.IDGenerator

	ctx := context.Background()
	for _, target := range f.Targets {
		if err := svc.PutTarget(ctx, target); err != nil {
			t.Fatalf("failed to populate scraper targets")
		}
	}

	handler := NewScraperHandler()
	handler.ScraperStorageService = svc
	server := httptest.NewServer(handler)
	client := ScraperService{
		Addr:     server.URL,
		OpPrefix: inmem.OpPrefix,
	}
	done := server.Close

	return &client, inmem.OpPrefix, done
}

func TestScraperService(t *testing.T) {
	platformtesting.ScraperService(initScraperService, t)
}