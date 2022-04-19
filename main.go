package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func CreateAccessKey(svc *iam.IAM, username string) *iam.AccessKey {
	result, err := svc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})

	if err != nil {
		log.Fatal("Error", err)
	}

	fmt.Printf("Create new Access Key for User: %s\n", username)
	return result.AccessKey
}

func ListUsers(svc *iam.IAM) []*iam.User {
	input := &iam.ListUsersInput{}
	result, err := svc.ListUsers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}

	}
	return result.Users
}

func ListAccessKeys(svc *iam.IAM, username string) *iam.ListAccessKeysOutput {
	result, err := svc.ListAccessKeys(&iam.ListAccessKeysInput{
		MaxItems: aws.Int64(5),
		UserName: aws.String(username),
	})

	if err != nil {
		log.Fatal("Error", err)
	}

	return result
}

func GetAccessKeyLastUsed(svc *iam.IAM, acces_key_id string) *time.Time {
	result, err := svc.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{
		AccessKeyId: aws.String(acces_key_id),
	})

	if err != nil {
		log.Fatal("Error", err)
	}

	return result.AccessKeyLastUsed.LastUsedDate
}

func UpdateAccessKeyStatus(svc *iam.IAM, username, acces_key_id string) {
	_, err := svc.UpdateAccessKey(&iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(acces_key_id),
		Status:      aws.String(iam.StatusTypeInactive),
		UserName:    aws.String(username),
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Access Key %s status updated to: Inactive for %s\n", acces_key_id, username)
}

func DeleteAccessKey(svc *iam.IAM, username, acces_key_id string) {
	_, err := svc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(acces_key_id),
		UserName:    aws.String(username),
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Printf("Successfully deleted Access Key %s for %s\n", acces_key_id, username)
}

func get_older_access_key(num_access_keys int, access_keys []*iam.AccessKeyMetadata) string {
	if num_access_keys == 1 {
		return *access_keys[0].AccessKeyId
	}

	access_key_one := access_keys[0]
	access_key_two := access_keys[1]
	// Compare the existing two access key's creation date, and returns the older one
	if access_key_one.CreateDate.Sub(*access_key_two.CreateDate).Round(1*time.Minute).Minutes() < 0 {
		return *access_key_one.AccessKeyId
	}
	return *access_key_two.AccessKeyId
}

func init_process(svc *iam.IAM, process, username string, accessKeys *iam.ListAccessKeysOutput) {
	// Maximum number of Access Keys per User
	max_num_of_access_keys := 2
	// Get number of access keys
	num_access_keys := len(accessKeys.AccessKeyMetadata)
	// Get older access key id
	older_access_key_id := get_older_access_key(num_access_keys, accessKeys.AccessKeyMetadata)

	if process == "create" {
		if num_access_keys == max_num_of_access_keys {
			// If user has two access keys, delete the older one first, then create a new one
			DeleteAccessKey(svc, username, older_access_key_id)
		}
		// Create a new access key
		accessKey := CreateAccessKey(svc, username)
		SendNewAccessKeyCredentials(accessKey, username)
	} else if process == "update" && num_access_keys == max_num_of_access_keys {
		// Inactivate old access key
		UpdateAccessKeyStatus(svc, username, older_access_key_id)
		SendNofitication(older_access_key_id, process, username)
	} else if process == "delete" && num_access_keys == max_num_of_access_keys {
		// Delete older/inactive access key
		DeleteAccessKey(svc, username, older_access_key_id)
		SendNofitication(older_access_key_id, process, username)
	}
}

func valid_process_name(process string) bool {
	valid_processes := []string{"create", "update", "delete"}
	for _, valid_process_name := range valid_processes {
		if process == valid_process_name {
			return true
		}
	}
	return false
}

func exec_process(usernames []*iam.User, svc *iam.IAM, process string) {
	// Filter only users with the given pattern for key rotation
	// For example to list only IL users set USERNAME_PATTERN to "@infinitelambda.com"
	username_pattern := os.Getenv("USERNAME_PATTERN")

	for _, user := range usernames {
		username := *user.UserName
		if strings.Contains(username, username_pattern) {
			// Get User's access keys
			accessKeys := ListAccessKeys(svc, username)
			// Initialize a process depending on the pre-defined variable
			init_process(svc, process, username, accessKeys)
		}
	}
}

func main() {
	// The PROCESS will come from Gitlab's CI/CD variable depending on the actual pipeline
	input_process := os.Getenv("PROCESS")
	// Convert input_process variable into a lowercase word
	process := strings.ToLower(input_process)

	// Valid processes are: "create", "update", "delete"
	if !valid_process_name(process) {
		log.Fatal("Please specify a valid process.\nValid processes are: create, update, delete")
	}

	aws_region := os.Getenv("AWS_REGION")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(aws_region)},
	)

	if err != nil {
		fmt.Println(err)
	}

	// Create an IAM service client
	svc := iam.New(sess)

	// Get IAM users
	usernames := ListUsers(svc)

	exec_process(usernames, svc, process)
}
