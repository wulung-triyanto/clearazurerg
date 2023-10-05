package auto

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

func getSubscriptionId() (string, error) {
	sub, ok := os.LookupEnv("SUBSCRIPTION_ID")

	if !ok {
		return "", errors.New("Azure subscription id not found.")
	}

	return sub, nil
}

func applyUpdate(updates []string, client *armresources.ResourceGroupsClient) error {
	for _, u := range updates {
		r, err := client.Update(context.Background(), u, armresources.ResourceGroupPatchable{
			Tags: map[string]*string{
				getExpirationTagName(): to.Ptr(getExpiration()),
			}}, nil)

		if err != nil {
			log.Printf("Error while updating Resource Group '%s': %s", u, err)
			return err
		}

		log.Printf("Resource Group '%s' updated successfully", *r.Name)
	}

	return nil
}

func applyRemoval(removals []string, client *armresources.ResourceGroupsClient) error {
	for _, r := range removals {
		p, err := client.BeginDelete(context.Background(), r, nil)

		if err != nil {
			log.Printf("Error while trying to begin removal of Resource Group '%s'", r)
			return err
		}

		option := new(runtime.PollUntilDoneOptions)
		option.Frequency = 15 * time.Second
		_, err = p.PollUntilDone(context.Background(), option)
		if err != nil {
			log.Printf("Error while removing Resource Group '%s'", r)
			return err
		}
	}

	return nil
}

func CleanUpResourceGroups() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Printf("Error while loading Azure credentials %s", err)
		return err
	}
	s, err := getSubscriptionId()
	if err != nil {
		return err
	}

	client, err := armresources.NewResourceGroupsClient(s, cred, nil)

	for pager := client.NewListPager(nil); pager.More(); {
		pager, err := pager.NextPage(context.Background())
		if err != nil {
			log.Printf("Failed to load next page %s", err)
			return err
		}

		updates := make([]string, 0)
		removals := make([]string, 0)

		for _, rg := range pager.ResourceGroupListResult.Value {
			log.Printf("\nLooking at Resource Group %s\n", *rg.Name)

			if hasKeeperTag(rg.Tags) {
				log.Printf("Resource Group '%s' has keeper tag. Will leave it as it is.", *rg.Name)
				continue
			} else if !hasExpirationTag(rg.Tags) {
				log.Printf("Resource Group '%s' will be marked for expiration.", *rg.Name)
				updates = append(updates, *rg.Name)
				continue
			} else if isExpired(rg.Tags) {
				log.Printf("Resource Group '%s' is expired... Will delete it.", *rg.Name)
				removals = append(removals, *rg.Name)
				continue
			} else {
				log.Printf("Resource Group '%s' already marked for expiration. Will check again at next run...", *rg.Name)
			}
		}

		applyUpdate(updates, client)
		applyRemoval(removals, client)
	}

	return nil
}

func HandleTick(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.NotFound(writer, request)
		return
	}

	if err := CleanUpResourceGroups(); err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
