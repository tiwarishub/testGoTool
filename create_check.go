package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"storagecache"

	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// main function to run gosdk tests.
// creates azure managed service Identity client from the Environment Variables.
// sets timeout for mock tests of 120 seconds per call
// runs create, get, and delete for mock caches.

func main() {
	var azureSubscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	var azureResourceGroup = os.Getenv("RESOURCE_GROUP_NAME")
	var azureSubnet = os.Getenv("SUBNET_NAME")
	var azureLocation = os.Getenv("_REGION")

	// create a cache client
	cacheClient := storagecache.NewCachesClient(azureSubscriptionID)
	// create an authorizer from env vars or Azure Managed Service Idenity
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		cacheClient.Authorizer = authorizer
	}
	if err != nil {
		fmt.Println("cannot get authorizer")
		fmt.Printf("Error: '%s': /n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	//initialize parameters for create cache
	var size = int32(3072)
	var location string
	var subnet string
	var cacheName string
	var skuType string

	location = azureLocation
	subnet = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/vnet_%s/subnets/%s", azureSubscriptionID, azureResourceGroup, azureResourceGroup, azureSubnet)
	cacheName = "goSDK"
	skuType = "Standard_2G"

	var sku = storagecache.CacheSku{
		Name: &skuType}

	var cacheProperties = storagecache.CacheProperties{
		CacheSizeGB: &size,
		Subnet:      &subnet}

	var cacheCreate = storagecache.Cache{
		Location:        &location,
		Name:            &cacheName,
		CacheProperties: &cacheProperties,
		Sku:             &sku}

	cache, errCacheCreate := cacheClient.CreateOrUpdate(ctx,
		azureResourceGroup,
		cacheName,
		&cacheCreate)

	if errCacheCreate != nil {
		fmt.Println("cannot get the cache to complete create or update response")
		fmt.Printf("Error: '%s': /n", errCacheCreate)
		fmt.Println("##vso[build.addbuildtag]FAIL put cache GoSDK")
		fmt.Println("##vso[task.complete result=Failed;]")
		os.Exit(1)
	}
	fmt.Printf("'%s' Cache details \n", cache)
	fmt.Println("Going into wait for create cache completion")

	waitCache := cache.WaitForCompletionRef(ctx, cacheClient.Client)

	if waitCache != nil {
		fmt.Println("cannot get the cache to complete create or update response")
		fmt.Println("##vso[build.addbuildtag]FAIL get cache GoSDK")
		fmt.Println("##vso[task.complete result=Failed;]")
		os.Exit(1)
	}

	fmt.Println("Going into get cache call")
	cacheGet, errGetCache := cacheClient.Get(ctx,
		azureResourceGroup,
		cacheName)

	res, _ := json.Marshal(cacheGet)

	if errGetCache != nil {
		fmt.Println("Error during Get Call")
		fmt.Printf("Error: '%s': /n", errGetCache)
		fmt.Println("##vso[build.addbuildtag]FAIL get cache GoSDK")
		fmt.Println("##vso[task.complete result=Failed;]")
		os.Exit(1)
	}
	fmt.Println(string(res))
	fmt.Println("Going into Delete Cache call")

	// TODO: add storage target create/get/delete
	cacheDelete, errDeleteCache := cacheClient.Delete(ctx,
		azureResourceGroup,
		cacheName)

	responseDelete, _ := json.Marshal(cacheDelete)

	if errDeleteCache != nil {
		fmt.Println("Error during initial Delete Call")
		fmt.Printf("Error: '%s': /n", errDeleteCache)
		fmt.Println("##vso[build.addbuildtag]FAIL delete cache GoSDK")
		fmt.Println("##vso[task.complete result=Failed;]")
		os.Exit(1)
	}
	fmt.Println(string(responseDelete))

	waitCacheDelete := cache.WaitForCompletionRef(ctx, cacheClient.Client)

	if waitCacheDelete != nil {
		fmt.Println("Cache Delete Failed")
		fmt.Println("##vso[build.addbuildtag]FAIL delete cache GoSDK")
		fmt.Println("##vso[task.complete result=Failed;]")
		os.Exit(1)
	}
	fmt.Println("Cache delete finished - tests are done")

}
