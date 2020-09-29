package pulsarmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/skulup/operator-pkg/k8s/secret"
	"github.com/skulup/operator-pkg/reconciler"
	"github.com/skulup/pulsar-operator/api/v1alpha1"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
)

const defaultSuperuserPasswordLength = 10
const (
	secretUsernameField     = "username"
	secretPasswordField     = "password"
	secretEmailAddressField = "email"
)

func reconcileSecretes(ctx reconciler.Context, manager *v1alpha1.PulsarManager) error {
	sec := &v1.Secret{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: superUserSecretNamespace(manager),
		Name:      superUserSecretName(manager),
	}, sec,
		func() (err error) {
			// Secret found
			return
		},
		func() (err error) {
			// Secret not found
			if sec, err = createSuperUserSecret(manager); err == nil {
				if err = ctx.SetOwnershipReference(manager, sec); err == nil {
					if err = ctx.Client().Create(context.TODO(), sec); err == nil {
						ctx.Logger().Info("Secret creation success. ",
							"Secret.Name", sec.GetName(),
							"Secret.Namespace", sec.GetNamespace())
					}
				}
			}
			return
		})
}

func reconcileSuperUserAccount(ctx reconciler.Context, manager *v1alpha1.PulsarManager) error {
	sec := &v1.Secret{}
	return ctx.GetResource(types.NamespacedName{
		Namespace: superUserSecretNamespace(manager),
		Name:      superUserSecretName(manager),
	},
		sec,
		func() (err error) {
			var body []byte
			var resp *http.Response
			// Secret found
			managerBackendUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:7750/pulsar-manager",
				serviceName(manager), serviceNamespace(manager))
			if resp, err = http.Get(managerBackendUrl + "/csrf-token"); err == nil {
				defer resp.Body.Close()
				if body, err = ioutil.ReadAll(resp.Body); err == nil {
					csrfToken := string(body)
					if body, err = json.Marshal(map[string]string{
						"email":       string(sec.Data[secretEmailAddressField]),
						"name":        string(sec.Data[secretUsernameField]),
						"password":    string(sec.Data[secretPasswordField]),
						"description": "super-user-account",
					}); err == nil {
						var req *http.Request
						if req, err = http.NewRequest("PUT", managerBackendUrl+"/users/superuser", bytes.NewBuffer(body)); err == nil {
							req.Header.Set("X-XSRF-TOKEN", csrfToken)
							req.Header.Set("Content-Type", "application/json")
							req.Header.Set("Cookie", fmt.Sprintf("XSRF-TOKEN=%s;", csrfToken))
							if resp, err = http.DefaultClient.Do(req); err == nil {
								defer resp.Body.Close()
								ctx.Logger().Info("Super super account creation success. ",
									"PulsarManager.Name", manager.GetName(),
									"PulsarManager.Namespace", manager.GetNamespace())
							}
						}
					}

				}
			}
			return
		},
		func() (err error) {
			// Secret not found
			return
		})
}

func createSuperUserSecret(manager *v1alpha1.PulsarManager) (*v1.Secret, error) {
	superUserPassword, err := secret.NewPassword(defaultSuperuserPasswordLength)
	if err != nil {
		return nil, err
	}
	return secret.New(
		superUserSecretNamespace(manager),
		superUserSecretName(manager),
		map[string][]byte{
			secretUsernameField:     []byte(manager.Spec.Username),
			secretEmailAddressField: []byte(manager.Spec.Email),
			secretPasswordField:     []byte(superUserPassword),
		}), nil
}

func superUserSecretNamespace(manager *v1alpha1.PulsarManager) string {
	return manager.GetNamespace()
}

func superUserSecretName(manager *v1alpha1.PulsarManager) string {
	return fmt.Sprintf("%s-pulsar-manager-super-user", manager.GetName())
}
