// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package minimal_downtime_migration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	// "os"
	"testing"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/common/constants"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/internal"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/profiles"
	// "github.com/GoogleCloudPlatform/spanner-migration-tool/testing/minimal_downtime_migration/util"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/webv2"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/webv2/helpers"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/webv2/session"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/logger"
	"github.com/google/uuid"
)

type migrationDetails struct {
	TargetDetails    targetDetails             `json:"TargetDetails"`
	DatastreamConfig profiles.DatastreamConfig `json:"DatastreamConfig"`
	GcsConfig        profiles.GcsConfig        `json:"GcsConfig"`
	DataflowConfig   profiles.DataflowConfig   `json:"DataflowConfig"`
	MigrationMode    string                    `json:"MigrationMode"`
	MigrationType    string                    `json:"MigrationType"`
	IsSharded        bool                      `json:"IsSharded"`
	SkipForeignKeys  bool                      `json:"skipForeignKeys"`
}

type targetDetails struct {
	TargetDB                    string `json:"TargetDB"`
	SourceConnectionProfileName string `json:"SourceConnProfile"`
	TargetConnectionProfileName string `json:"TargetConnProfile"`
	ReplicationSlot             string `json:"ReplicationSlot"`
	Publication                 string `json:"Publication"`
}

func TestSchemaAndDataMySQLToSpannerMigration(t *testing.T) {
	sessionState := session.GetSessionState()
	logger.InitializeLogger("info")

	sessionState.Driver = constants.MYSQL
	sessionState.Conv = internal.MakeConv()
	projectID, err := metadata.ProjectID()
	if err != nil {
		t.Fatal(err)
	}
	projectID = "span-cloud-testing"
	sessionState.GCPProjectID = "span-cloud-testing"
	sessionState.SpannerInstanceID = "djagaluru-dms-test"
	sessionState.SpannerDatabaseName = "testDb"
	sessionState.Dialect = constants.DIALECT_GOOGLESQL
	sessionState.Region = "us-central1"
	sessionState.Bucket = "djagaluru-test-datastream"
	sessionState.RootPath = "temp-1234"

	sessionFile := "test_data/mysql_session_test.json"

	// read session and parse to session object
	s, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(s, &sessionState.Conv)
	if err != nil {
		t.Fatal(err)
	}

	sessionState.SourceDBConnDetails = session.SourceDBConnDetails{
		ConnectionType: helpers.DIRECT_CONNECT_MODE,
		Host:     "34.69.106.17", //util.GetIp(),
		Port:     "3306",
		User:     "root", // os.Getenv("MYSQLUSER"),
		Password: "root", // os.Getenv("MYSQLPWD"),
	}
	sessionState.DbName = "l1"
	sessionState.Conv.Audit.MigrationRequestId = uuid.NewString()

	payload := migrationDetails{
		IsSharded: false,
		MigrationType: helpers.LOW_DOWNTIME_MIGRATION,
		TargetDetails: targetDetails{
			TargetDB: "migration_test_1234",
			TargetConnectionProfileName: "djagaluru-gen-data-destination", // "target-profile-id",
			SourceConnectionProfileName: "djagaluru-gen-data", //"source-profile-id",
		},
		DatastreamConfig: profiles.DatastreamConfig{MaxConcurrentBackfillTasks: "10", MaxConcurrentCdcTasks: "10"},
		DataflowConfig: profiles.DataflowConfig{
			ProjectId: projectID,
			Location: "us-central1",
		},
	}

	inputBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(inputBytes)

	req, err := http.NewRequest("POST", "/Migrate", buffer)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webv2.Migrate)
	handler.ServeHTTP(rr, req)
	var result map[string]string
	json.Unmarshal(rr.Body.Bytes(), &result)
	time.Sleep(10 * time.Minute)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}