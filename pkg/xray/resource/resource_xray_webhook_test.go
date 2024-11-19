package xray_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-xray/v3/pkg/acctest"
)

func TestAccWebhook_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("webhook", "xray_webhook")
	url := fmt.Sprintf("https://tempurl%d.org", testutil.RandomInt())

	const template = `
		resource "xray_webhook" "{{ .name }}" {
			name        = "{{ .name }}"
			description = "{{ .description }}"
			url         = "{{ .url }}"
			use_proxy   = {{ .use_proxy }}
			user_name   = "{{ .user_name }}"
			password    = "{{ .password }}"

			headers = {
				{{ .header1_name }} = "{{ .header1_value }}"
				{{ .header2_name }} = "{{ .header2_value }}"
			}
		}
	`
	testData := map[string]string{
		"name":          resourceName,
		"description":   "test description",
		"url":           url,
		"use_proxy":     "true",
		"user_name":     "test_user_1",
		"password":      "test_password_1",
		"header1_name":  "header1_name",
		"header1_value": "header1_value",
		"header2_name":  "header2_name",
		"header2_value": "header2_value",
	}

	config := util.ExecuteTemplate("TestAccWebhook_UpgradeFromSDKv2", template, testData)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"xray": {
						VersionConstraint: "2.4.0",
						Source:            "jfrog/xray",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "description", testData["description"]),
					resource.TestCheckResourceAttr(fqrn, "url", testData["url"]),
					resource.TestCheckResourceAttr(fqrn, "use_proxy", testData["use_proxy"]),
					resource.TestCheckResourceAttr(fqrn, "user_name", testData["user_name"]),
					resource.TestCheckResourceAttr(fqrn, "password", testData["password"]),
					resource.TestCheckResourceAttr(fqrn, "headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "headers.header1_name", testData["header1_value"]),
					resource.TestCheckResourceAttr(fqrn, "headers.header2_name", testData["header2_value"]),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks:         testutil.ConfigPlanChecks(""),
			},
		},
	})
}

func TestAccWebhook_full(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("webhook", "xray_webhook")
	url := fmt.Sprintf("https://tempurl%d.org", testutil.RandomInt())

	const template = `
		resource "xray_webhook" "{{ .name }}" {
			name        = "{{ .name }}"
			description = "{{ .description }}"
			url         = "{{ .url }}"
			use_proxy   = {{ .use_proxy }}
			user_name   = "{{ .user_name }}"
			password    = "{{ .password }}"

			headers = {
				{{ .header1_name }} = "{{ .header1_value }}"
				{{ .header2_name }} = "{{ .header2_value }}"
			}
		}
	`
	testData := map[string]string{
		"name":          resourceName,
		"description":   "test description",
		"url":           url,
		"use_proxy":     "true",
		"user_name":     "test_user_1",
		"password":      "test_password_1",
		"header1_name":  "header1_name",
		"header1_value": "header1_value",
		"header2_name":  "header2_name",
		"header2_value": "header2_value",
	}

	config := util.ExecuteTemplate("TestAccWebhook_full", template, testData)

	const updateTemplate = `
		resource "xray_webhook" "{{ .name }}" {
			name        = "{{ .name }}"
			description = "{{ .description }}"
			url         = "{{ .url }}"
			use_proxy   = "{{ .use_proxy }}"
		}
	`
	updatedTestData := map[string]string{
		"name":        resourceName,
		"description": "test description 2",
		"url":         url,
		"use_proxy":   "false",
	}
	updatedConfig := util.ExecuteTemplate("TestAccWebhook_full", updateTemplate, updatedTestData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "name", testCheckWebhook),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "description", testData["description"]),
					resource.TestCheckResourceAttr(fqrn, "url", testData["url"]),
					resource.TestCheckResourceAttr(fqrn, "use_proxy", testData["use_proxy"]),
					resource.TestCheckResourceAttr(fqrn, "user_name", testData["user_name"]),
					resource.TestCheckResourceAttr(fqrn, "password", testData["password"]),
					resource.TestCheckResourceAttr(fqrn, "headers.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "headers.header1_name", testData["header1_value"]),
					resource.TestCheckResourceAttr(fqrn, "headers.header2_name", testData["header2_value"]),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", updatedTestData["name"]),
					resource.TestCheckResourceAttr(fqrn, "description", updatedTestData["description"]),
					resource.TestCheckResourceAttr(fqrn, "url", updatedTestData["url"]),
					resource.TestCheckResourceAttr(fqrn, "use_proxy", updatedTestData["use_proxy"]),
					resource.TestCheckNoResourceAttr(fqrn, "user_name"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckNoResourceAttr(fqrn, "headers.%"),
				),
				ConfigPlanChecks: testutil.ConfigPlanChecks(""),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        resourceName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

func TestAccWebhook_invalid_name(t *testing.T) {
	_, _, resourceName := testutil.MkNames("test-webhook", "xray_webhook")
	url := fmt.Sprintf("https://tempurl%d.org", testutil.RandomInt())

	const template = `
		resource "xray_webhook" "{{ .name }}" {
			name = "{{ .name }}"
			url  = "{{ .url }}"
		}
	`
	testData := map[string]string{
		"name": resourceName,
		"url":  url,
	}

	config := util.ExecuteTemplate("TestAccWebhook_invalid_name", template, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(".*must contain only alphanumberic characters.*"),
			},
		},
	})
}

func TestAccWebhook_invalid_url(t *testing.T) {
	_, _, resourceName := testutil.MkNames("webhook", "xray_webhook")

	const template = `
		resource "xray_webhook" "{{ .name }}" {
			name = "{{ .name }}"
			url  = "tempurl.org"
		}
	`
	testData := map[string]string{
		"name": resourceName,
	}

	config := util.ExecuteTemplate("TestAccWebhook_invalid_name", template, testData)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`.*Attribute url value must be a valid URL with host and http or https scheme.*`),
			},
		},
	})
}

func testCheckWebhook(id string, request *resty.Request) (*resty.Response, error) {
	return request.
		SetPathParam("id", id).
		Get("xray/api/v1/webhooks/{id}")
}
