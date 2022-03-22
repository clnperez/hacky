package main

// usage: ./main "useridstringnotyouremailaddr" "lon06" "serviceidstringnotthecrn"
// hints: if you're lazy just put your e-mail address in, set debug to true, and the crn that is spit out will include your actual userid

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/IBM/go-sdk-core/v5/core"
	powervsproviderv1 "github.com/openshift/machine-api-provider-powervs/pkg/apis/powervsprovider/v1alpha1"

	// powervsclient "github.com/openshift/machine-api-provider-powervs/pkg/client"
	goclient "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Specify ID. Zone, Cloud Inst ID")
		return
	}

	id := os.Args[1]
	zone := os.Args[2]
	cloudInstID := os.Args[3]
	if id == "" || zone == "" || cloudInstID == "" {
		fmt.Println("Specify ID. Zone, Cloud Inst ID")
		return
	}

	apikey := os.Getenv("IBMCLOUD_API_KEY")
	if apikey == "" {
		fmt.Println("empty IBMCLOUD_API_KEY env var")
		return
	}

	ctx := context.Background()
	authenticator := &core.IamAuthenticator{
		ApiKey: apikey,
	}

	o := &ibmpisession.IBMPIOptions{
		Authenticator: authenticator,
		UserAccount:   id,
		Zone:          zone,
		//Debug:         true,
	}

	s, err := ibmpisession.NewIBMPISession(o)
	if err != nil {
		fmt.Println(err)
		return
	}

	nwclient := goclient.NewIBMPINetworkClient(ctx, s, cloudInstID)

	network := powervsproviderv1.PowerVSResourceReference{}

	if id, err := getNetworkID(network, nwclient); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(*id)
	}
}

func getNetworkID(network powervsproviderv1.PowerVSResourceReference, client *goclient.IBMPINetworkClient) (*string, error) {
	if network.ID != nil {
		return network.ID, nil
	} else {
		networks, err := client.GetAll()
		if err != nil {
			return nil, err
		}
		if network.Name != nil {
			for _, nw := range networks.Networks {
				if *network.Name == *nw.Name {
					return nw.NetworkID, nil
				}
			}
		} else {
			for _, nw := range networks.Networks {
				match, err := regexp.Match("^DHCPSERVER[0-9a-z]{32}_Private$", []byte(*nw.Name))
				if err != nil {
					return nil, err
				} else if match {
					return nw.NetworkID, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("failed to find a network ID")
}
