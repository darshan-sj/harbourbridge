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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/common/constants"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/common/utils"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/internal"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/testing/minimal_downtime_migration/util"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/webv2/profile"
	"github.com/GoogleCloudPlatform/spanner-migration-tool/webv2/session"
)

func TestCreateConnectionProfileMySQL(t *testing.T) {
	sessionState := session.GetSessionState()

	sessionState.Driver = constants.MYSQL
	sessionState.Conv = internal.MakeConv()
	projectID, err := metadata.ProjectID()
	if err != nil {
		t.Fatal(err)
	}
	sessionState.GCPProjectID = projectID
	sessionState.Region = "us-central1"
	sessionState.SourceDBConnDetails = session.SourceDBConnDetails{
		Host:     util.GetIp(),
		Port:     "3306",
		User:     os.Getenv("MYSQLUSER"),
		Password: os.Getenv("MYSQLPWD"),
	}
	id, err := utils.GenerateName("it-source-")
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]interface{}{
		"Id":           id,
		"IsSource":     true,
		"ValidateOnly": false,
	}

	inputBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(inputBytes)

	req, err := http.NewRequest("POST", "/CreateConnectionProfile", buffer)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(profile.CreateConnectionProfile)
	handler.ServeHTTP(rr, req)
	var result map[string]string
	json.Unmarshal(rr.Body.Bytes(), &result)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	connectionProfileID := fmt.Sprintf("projects/%s/locations/%s/connectionProfiles/%s", projectID, "us-central1", id)

	datastreamRM := util.DSResourceManager{}
	connectionProfile, err := datastreamRM.GetConnectionProfile(connectionProfileID)
	if err != nil {
		t.Errorf("Could not fetch created connection profile: err=%v", err)
	}
	if connectionProfile.GetName() != connectionProfileID {
		t.Errorf("Wrong connection profile: got %v want %v", connectionProfile.GetName(), fmt.Sprintf("projects/%s/locations/%s/connectionProfiles/%s", projectID, "us-central1", id))
	}
	err = datastreamRM.DeleteConnectionProfile(connectionProfileID)
	if err != nil {
		t.Errorf("Could not cleanup connection profile: err=%v", err)
	}
}
