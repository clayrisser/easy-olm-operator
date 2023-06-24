package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func GetCrd(ctx context.Context, c client.Client, crd string) (*unstructured.Unstructured, error) {
	if _, err := url.ParseRequestURI(crd); err == nil {
		crd, err = fetchCrd(crd)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch CRD from URL: %w", err)
		}
	}
	crdObj, err := getCrd(ctx, c, crd)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRD: %w", err)
	}
	return crdObj, nil
}

func getCrd(ctx context.Context, c client.Client, crd string) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	jsonSpec, err := yaml.YAMLToJSON([]byte(crd))
	if err != nil {
		return nil, err
	}
	if err := obj.UnmarshalJSON(jsonSpec); err != nil {
		return nil, err
	}
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Kind != "CustomResourceDefinition" || gvk.Group != "apiextensions.k8s.io" {
		return nil, errors.New("the provided YAML is not a CRD")
	}
	if err := c.Create(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func fetchCrd(crdUrl string) (string, error) {
	resp, err := http.Get(crdUrl)
	if err != nil {
		return "", fmt.Errorf("failed to send a GET request to the CRD URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the HTTP response body: %w", err)
	}
	return string(body), nil
}
