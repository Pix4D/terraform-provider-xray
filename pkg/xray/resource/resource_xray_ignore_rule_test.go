package xray_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-xray/v3/pkg/acctest"
)

func TestAccIgnoreRule_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  artifact {
			  name    = "fake-name"
			  version = "fake-version"
			  path    = "fake-path/"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"xray": {
						VersionConstraint: "2.8.1",
						Source:            "jfrog/xray",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "artifact.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.name", "fake-name"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.version", "fake-version"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.path", "fake-path/"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
			},
		},
	})
}

func TestAccIgnoreRule_objectives(t *testing.T) {
	for _, objective := range []string{"vulnerabilities", "cves", "licenses"} {
		t.Run(objective, func(t *testing.T) {
			resource.Test(objectiveTestCase(objective, t))
		})
	}
}

func objectiveTestCase(objective string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  {{ .objective }} = ["fake-{{ .objective }}"]
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
		"objective":      objective,
	})

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, fmt.Sprintf("%s.#", objective), "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, fmt.Sprintf("%s.*", objective), fmt.Sprintf("fake-%s", objective)),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	}
}

func TestAccIgnoreRule_operational_risk(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  operational_risk = ["any"]

  		  component {
		    name = "fake-component"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "operational_risk.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "operational_risk.*", "any"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_invalid_operational_risk(t *testing.T) {
	_, _, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  operational_risk = ["invalid-risk"]

  		  component {
		    name = "fake-component"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`.*Attribute operational_risk\[Value\(".+"\)\] value must be one of:.*`),
			},
		},
	})
}

func TestAccIgnoreRule_scopes_policies(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_security_policy" "fake_policy" {
			name        = "fake-policy"
			description = "Security policy description"
			type        = "security"
			rule {
				name     = "rule-name-severity"
				priority = 1
				criteria {
				min_severity = "High"
			}
			actions {
				mails    = ["test@email.com"]
				block_download {
					unscanned = true
					active    = true
				}
				block_release_bundle_distribution  = true
				fail_build                         = true
				notify_watch_recipients            = true
				notify_deployer                    = true
				create_ticket_enabled              = false
				build_failure_grace_period_in_days = 5
				}
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
			notes            = "fake notes"
			expiration_date  = "{{ .expirationDate }}"
			cves             = ["fake-cve"]
		 	policies         = [xray_security_policy.fake_policy.name]
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "policies.*", "fake-policy"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_scopes_watches_policies(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_security_policy" "security" {
			name        = "fake-policy"
			description = "Security policy description"
			type        = "security"
			rule {
				name     = "rule-name-severity"
				priority = 1
				criteria {
				min_severity = "High"
			}
			actions {
				mails    = ["test@email.com"]
				block_download {
					unscanned = true
					active    = true
				}
				block_release_bundle_distribution  = true
				fail_build                         = true
				notify_watch_recipients            = true
				notify_deployer                    = true
				create_ticket_enabled              = false
				build_failure_grace_period_in_days = 5
				}
			}
		}
		resource "xray_watch" "fake_watch" {
			name          = "fake-watch"
			active 		  = true

			watch_resource {
				type       	= "all-repos"
				filter {
					type  	= "regex"
					value	= ".*"
				}
			}
			assigned_policy {
				name = xray_security_policy.security.name
				type = "security"
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
			notes            = "fake notes"
			expiration_date  = "{{ .expirationDate }}"
			cves             = ["fake-cve"]
		 	watches          = [xray_watch.fake_watch.name]
			policies         = [xray_security_policy.security.name]
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "watches.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watches.*", "fake-watch"),
					resource.TestCheckResourceAttr(fqrn, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "policies.*", "fake-policy"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_scopes_no_expiration_policies(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_security_policy" "security" {
			name        = "fake-policy"
			description = "Security policy description"
			type        = "security"
			rule {
				name     = "rule-name-severity"
				priority = 1
				criteria {
					min_severity = "High"
				}
				actions {
					mails    = ["test@email.com"]
					block_download {
						unscanned = true
						active    = true
					}
					block_release_bundle_distribution  = true
					fail_build                         = true
					notify_watch_recipients            = true
					notify_deployer                    = true
					create_ticket_enabled              = false
					build_failure_grace_period_in_days = 5
				}
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
		  notes    = "fake notes"
		  cves     = ["fake-cve"]
		  policies = [xray_security_policy.security.name]
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckNoResourceAttr(fqrn, "expiration_date"),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "policies.*", "fake-policy"),
				),
			},
		},
	})
}

func TestAccIgnoreRule_scopes_no_expiration_watches(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_security_policy" "security" {
			name        = "fake-policy"
			description = "Security policy description"
			type        = "security"
			rule {
				name     = "rule-name-severity"
				priority = 1
				criteria {
					min_severity = "High"
				}
				actions {
					mails    = ["test@email.com"]
					block_download {
						unscanned = true
						active    = true
					}
					block_release_bundle_distribution  = true
					fail_build                         = true
					notify_watch_recipients            = true
					notify_deployer                    = true
					create_ticket_enabled              = false
					build_failure_grace_period_in_days = 5
				}
			}
		}

		resource "xray_watch" "fake_watch" {
			name   = "fake-watch"
			active = true

			watch_resource {
				type      = "all-repos"
				filter {
					type  = "regex"
					value = ".*"
				}
			}
			assigned_policy {
				name = xray_security_policy.security.name
				type = "security"
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
		  notes   = "fake notes"
		  cves    = ["fake-cve"]
		  watches = [xray_watch.fake_watch.name]
		}
	`, map[string]interface{}{
		"name": name,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckNoResourceAttr(fqrn, "expiration_date"),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "watches.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watches.*", "fake-watch"),
				),
			},
		},
	})
}

func TestAccIgnoreRule_docker_layers(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  docker_layers = [
		    "2ae0e4835a9a6e22e35dd0fcce7d7354999476b7dad8698d2d7a77c80bfc647b",
			"a8db0e25d5916e70023114bb2d2497cd85327486bd6e0dc2092b349a1ab3a0a0"
		  ]
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "docker_layers.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "docker_layers.*", "2ae0e4835a9a6e22e35dd0fcce7d7354999476b7dad8698d2d7a77c80bfc647b"),
					resource.TestCheckTypeSetElemAttr(fqrn, "docker_layers.*", "a8db0e25d5916e70023114bb2d2497cd85327486bd6e0dc2092b349a1ab3a0a0"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_invalid_docker_layers(t *testing.T) {
	_, _, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]

		  docker_layers = [
		    "invalid-layer",
			"a8db0e25d5916e70023114bb2d2497cd85327486bd6e0dc2092b349a1ab3a0a0"
		  ]
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Must be SHA256 hash"),
			},
		},
	})
}

func TestAccIgnoreRule_sources(t *testing.T) {
	// can't easily test "release_bundle" as the xray instance for acceptance testing
	// doesn't have all the components (gpg key, mission control, etc.)
	for _, source := range []string{"build", "component"} {
		t.Run(source, func(t *testing.T) {
			resource.Test(sourceTestCase(source, t))
		})
	}
}

func sourceTestCase(source string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  {{ .source }} {
			  name    = "fake-name"
			  version = "fake-version"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
		"source":         source,
	})

	return t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, fmt.Sprintf("%s.#", source), "1"),
					resource.TestCheckResourceAttr(fqrn, fmt.Sprintf("%s.0.name", source), "fake-name"),
					resource.TestCheckResourceAttr(fqrn, fmt.Sprintf("%s.0.version", source), "fake-version"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	}
}

func TestAccIgnoreRule_artifact(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().UTC().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  artifact {
			  name    = "fake-name"
			  version = "fake-version"
			  path    = "fake-path/"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Local().Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Local().Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "artifact.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.name", "fake-name"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.version", "fake-version"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.path", "fake-path/"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_artifact_with_no_version(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().UTC().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  artifact {
			  name    = "fake-name"
			  path    = "fake-path/"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Local().Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Local().Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "artifact.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.name", "fake-name"),
					resource.TestCheckNoResourceAttr(fqrn, "artifact.0.version"),
					resource.TestCheckResourceAttr(fqrn, "artifact.0.path", "fake-path/"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIgnoreRule_invalid_artifact_path(t *testing.T) {
	_, _, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]

		  artifact {
			  name    = "fake-name"
			  version = "fake-version"
			  path    = "invalid-path"
		  }
		}
	`, map[string]interface{}{
		"name":           name,
		"expirationDate": expirationDate.Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Must end with a '/'"),
			},
		},
	})
}

func TestAccIgnoreRule_with_project_key(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)
	projectKey := fmt.Sprintf("testproj%d", testutil.RandomInt())

	config := util.ExecuteTemplate(
		"TestAccIgnoreRule",
		`resource "project" "{{ .projectKey }}" {
			key          = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
			notes            = "fake notes"
			expiration_date  = "{{ .expirationDate }}"
			project_key      = project.{{ .projectKey }}.key

			licenses = ["unknown"]

			docker_layers = [
				"2ae0e4835a9a6e22e35dd0fcce7d7354999476b7dad8698d2d7a77c80bfc647b",
				"a8db0e25d5916e70023114bb2d2497cd85327486bd6e0dc2092b349a1ab3a0a0"
			]
		}`,
		map[string]interface{}{
			"name":           name,
			"expirationDate": expirationDate.Format("2006-01-02"),
			"projectKey":     projectKey,
		},
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
			},
		},
	})
}

func TestAccIgnoreRule_build_with_project_key(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	expirationDate := time.Now().Add(time.Hour * 48)
	projectKey := fmt.Sprintf("testproj%d", testutil.RandomInt())

	config := util.ExecuteTemplate(
		"TestAccIgnoreRule",
		`resource "project" "{{ .projectKey }}" {
			key          = "{{ .projectKey }}"
			display_name = "{{ .projectKey }}"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
		}

		resource "xray_ignore_rule" "{{ .name }}" {
			notes            = "fake notes"
			expiration_date  = "{{ .expirationDate }}"
			project_key      = project.{{ .projectKey }}.key

			licenses = ["unknown"]

			build {
				name    = "fake-name"
				version = "fake-version"
			}
		}`,
		map[string]interface{}{
			"name":           name,
			"expirationDate": expirationDate.Format("2006-01-02"),
			"projectKey":     projectKey,
		},
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
			},
		},
	})
}

// TestAccIgnoreRule_release_bundle would only be successful when executed against Artifactory instance
// that has Distribution enabled (i.e. has edge node(s) configured)
func TestAccIgnoreRule_release_bundle(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if !strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is not a cloud instance.", jfrogURL)
	}

	_, fqrn, name := testutil.MkNames("ignore-rule-", "xray_ignore_rule")
	_, _, releaseBundleName := testutil.MkNames("test-release-bundle-v1", "distribution_release_bundle_v1")
	expirationDate := time.Now().UTC().Add(time.Hour * 48)

	config := util.ExecuteTemplate("TestAccIgnoreRule", `
		resource "distribution_release_bundle_v1" "{{ .releaseBundleName }}" {
			name = "{{ .releaseBundleName }}"
			version = "1.0.0"
			sign_immediately = false
			description = "Test description"

			release_notes = {
				syntax = "plain_text"
				content = "test release notes"
			}

			spec = {
				queries = [{
					aql = "items.find({ \"repo\" : \"example-repo-local\" })"
					query_name: "query-1"

					mappings = [{
						input = "original_repository/(.*)"
						output = "new_repository/$1"
					}]

					added_props = [{
						key = "test-key"
						values = ["test-value"]
					}]
					
					exclude_props_patterns = [
						"test-patterns"
					]
				}]
			}
		}
			
		resource "xray_ignore_rule" "{{ .name }}" {
		  notes            = "fake notes"
		  expiration_date  = "{{ .expirationDate }}"
		  vulnerabilities  = ["any"]
		  cves             = ["any"]

		  release_bundle {
			  name    = distribution_release_bundle_v1.{{ .releaseBundleName }}.name
			  version = distribution_release_bundle_v1.{{ .releaseBundleName }}.version
		  }
		}
	`, map[string]interface{}{
		"releaseBundleName": releaseBundleName,
		"name":              name,
		"expirationDate":    expirationDate.Local().Format("2006-01-02"),
	})

	resource.Test(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"distribution": {
				Source: "jfrog/distribution",
			},
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "", testCheckIgnoreRule),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "notes", "fake notes"),
					resource.TestCheckResourceAttr(fqrn, "expiration_date", expirationDate.Local().Format("2006-01-02")),
					resource.TestCheckResourceAttr(fqrn, "is_expired", "false"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.name", releaseBundleName),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.version", "1.0.0"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testCheckIgnoreRule fetches the supposingly deleted ignore rule and verify it has been deleted
// Xray applies soft delete to ignore rule and adds 'deleted_by' and 'deleted_at'
// fields to the payload after a rule is deleted
// Thus we check for the field's existence and return 404 error resp
func testCheckIgnoreRule(id string, request *resty.Request) (*resty.Response, error) {
	type PartialIgnoreRule struct {
		DeletedAt string `json:"deleted_at"`
		DeletedBy string `json:"deleted_by"`
	}

	partialRule := PartialIgnoreRule{}

	res, err := request.
		AddRetryCondition(client.NeverRetry).
		SetResult(&partialRule).
		SetPathParam("id", id).
		Get("xray/api/v1/ignore_rules/{id}")
	if err != nil {
		return res, err
	}
	if res.IsError() && res.StatusCode() != http.StatusNotFound {
		return res, fmt.Errorf("%s", res.String())
	}

	if len(partialRule.DeletedAt) > 0 {
		res.RawResponse.StatusCode = http.StatusNotFound // may be we should set http.StatusGone instead?
		return res, nil
	}

	return res, nil
}
