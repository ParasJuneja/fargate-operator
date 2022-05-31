package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	admissionv1 "k8s.io/api/admission/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type spec struct {
	subnets             []string
	podExecutionRoleArn string
	selectors           []types.FargateProfileSelector
	tags                map[string]string
}

type FargateProfile struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	Spec               spec `json:"spec,omitempty"`
}

type FargateProfileList struct {
	meta_v1.TypeMeta   `json:"inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Items              []FargateProfile `json:"items"`
}

// func getNewFargateProfile(name string, namespace string, s spec) FargateProfile {
// 	fp := FargateProfile{

// 		spec: s,
// 	}
// 	return fp
// }

func (fp *FargateProfile) DeepCopyObject() runtime.Object {
	if fp == nil {
		return nil
	}
	fp_copy := new(FargateProfile)
	*fp_copy = *fp
	fp_copy.TypeMeta = fp.TypeMeta
	fp_copy.ObjectMeta = fp.ObjectMeta
	fp_copy.Spec = fp.Spec
	return fp_copy
}

func createFargateProfile(fp FargateProfile) (*eks.CreateFargateProfileOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client := eks.NewFromConfig(cfg)
	input := &eks.CreateFargateProfileInput{
		ClusterName:         aws.String(os.Getenv("environmentName") + "-EKS-Cluster"),
		FargateProfileName:  aws.String(fp.Name),
		PodExecutionRoleArn: aws.String(fp.Spec.podExecutionRoleArn),
		Selectors:           fp.Spec.selectors,
		Subnets:             fp.Spec.subnets,
		Tags:                fp.Spec.tags,
	}

	resp, err := client.CreateFargateProfile(context.TODO(), input)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func createFargateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var request admissionv1.AdmissionReview
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("JSON body in invalid format: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	if request.APIVersion != "admission.k8s.io/v1" || request.Kind != "AdmissionReview" {
		http.Error(w, fmt.Sprintf("wrong APIVersion or kind: %s - %s", request.APIVersion, request.Kind), http.StatusBadRequest)
		return
	}
	fmt.Printf("debug: %+v\n", request.Request)
	response := admissionv1.AdmissionReview{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     request.Request.UID,
			Allowed: true,
		},
	}

	var fp FargateProfile
	patchType := admissionv1.PatchTypeJSONPatch
	if request.Request.Kind.Group == "fop.io" && request.Request.Kind.Version == "v1" && request.Request.Kind.Kind == "FargateProfile" && request.Request.Operation == "CREATE" {
		patch := fmt.Sprintf(string(`[{"op": "add", "path": "/metadata/labels/fp.io\/createdBy", "value": %s}]`), request.Request.UserInfo.Username)
		response.Response.PatchType = &patchType
		response.Response.Patch = []byte(patch)

		decoder, err := admission.NewDecoder(runtime.NewScheme())
		if err != nil {
			fmt.Println("Unable to initialize decoder: ", err)
			return
		}
		decoder.DecodeRaw(request.Request.Object, &fp)
		resp, err := createFargateProfile(fp)
		if err != nil {
			fmt.Println("Unable to create FargateProfile: ", err)
			return
		}
		fmt.Println(resp)
	}

	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("JSON output marshal error: %s\n", err.Error()), http.StatusBadRequest)
		return
	}
	fmt.Printf("Got request, response: %s\n", string(out))

}


func deleteFargateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}