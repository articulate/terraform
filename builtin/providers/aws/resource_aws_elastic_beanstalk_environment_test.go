package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBeanstalkEnv_basic(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBeanstalkEnvConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("awseb_elastic_beanstalk_environment.tfenvtest", &app),
				),
			},
		},
	})
}

func testAccCheckBeanstalkEnvDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awseb_elastic_beanstalk_environment" {
			continue
		}

		// Try to find the environment
		describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
		if err == nil {
			switch {
			case len(resp.Environments) > 1:
				return fmt.Errorf("Error %d environments match, expected 1", len(resp.Environments))
			case len(resp.Environments) == 1:
				if *resp.Environments[0].Status == "Terminated" {
					return nil
				}
				return fmt.Errorf("Elastic Beanstalk ENV still exists.")
			default:
				return nil
			}
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidBeanstalkEnvID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckBeanstalkEnvExists(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn
		describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		}

		log.Printf("[DEBUG] Elastic Beanstalk Environment TEST describe opts: %s", describeBeanstalkEnvOpts)

		resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
		if err != nil {
			return err
		}
		if len(resp.Environments) == 0 {
			return fmt.Errorf("Elastic Beanstalk ENV not found.")
		}

		*app = *resp.Environments[0]

		return nil
	}
}

const testAccBeanstalkEnvConfig = `
resource "awseb_elastic_beanstalk_application" "tftest" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "awseb_elastic_beanstalk_environment" "tfenvtest" {
  name = "tf-test-name"
  application = "${awseb_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux 2015.09 v2.0.4 running Go 1.4"
  #solution_stack_name =
}
`
