// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"

	"github.com/kubernetes/dashboard/src/app/backend/plugin/apis/v1alpha1"
	fakePluginClientset "github.com/kubernetes/dashboard/src/app/backend/plugin/client/clientset/versioned/fake"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeK8sClient "k8s.io/client-go/kubernetes/fake"
)

var srcData = "randomPluginSourceCode"

func TestGetPluginSource(t *testing.T) {
	ns := "default"
	pluginName := "test-plugin"
	filename := "plugin-test.js"
	cfgMapName := "plugin-test-cfgMap"

	pcs := fakePluginClientset.NewSimpleClientset()
	cs := fakeK8sClient.NewSimpleClientset()

	_, err := GetPluginSource(pcs, cs, ns, pluginName)
	if err == nil {
		t.Errorf("error 'plugins.dashboard.k8s.io \"%s\" not found' did not occur", pluginName)
	}

	_, _ = pcs.DashboardV1alpha1().Plugins(ns).Create(&v1alpha1.Plugin{
		ObjectMeta: v1.ObjectMeta{Name: pluginName, Namespace: ns},
		Spec: v1alpha1.PluginSpec{
			Source: v1alpha1.Source{
				ConfigMapRef: &coreV1.ConfigMapEnvSource{
					LocalObjectReference: coreV1.LocalObjectReference{Name: cfgMapName},
				},
				Filename: filename}},
	})

	_, err = GetPluginSource(pcs, cs, ns, pluginName)
	if err == nil {
		t.Errorf("error 'configmaps \"%s\" not found' did not occur", cfgMapName)
	}

	_, _ = cs.CoreV1().ConfigMaps(ns).Create(&coreV1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: cfgMapName, Namespace: ns},
		Data: map[string]string{filename: srcData},
	})

	data, err := GetPluginSource(pcs, cs, ns, pluginName)
	if err != nil {
		t.Errorf("error while fetching plugin source: %s", err)
	}

	if !bytes.Equal(data, []byte(srcData)) {
		t.Error("bytes in configMap and bytes from GetPluginSource are different")
	}
}

func Test_servePluginSource(t *testing.T) {
	h := Handler{&fakeClientManager{}}

	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1//plugin/default/test-plugin", nil)
	req := restful.NewRequest(httpReq)

	httpWriter := httptest.NewRecorder()
	resp := restful.NewResponse(httpWriter)

	h.servePluginSource(req, resp)
}
