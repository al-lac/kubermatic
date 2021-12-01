// Code generated by go-swagger; DO NOT EDIT.

package gke

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewListGKEClusterSizesParams creates a new ListGKEClusterSizesParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListGKEClusterSizesParams() *ListGKEClusterSizesParams {
	return &ListGKEClusterSizesParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListGKEClusterSizesParamsWithTimeout creates a new ListGKEClusterSizesParams object
// with the ability to set a timeout on a request.
func NewListGKEClusterSizesParamsWithTimeout(timeout time.Duration) *ListGKEClusterSizesParams {
	return &ListGKEClusterSizesParams{
		timeout: timeout,
	}
}

// NewListGKEClusterSizesParamsWithContext creates a new ListGKEClusterSizesParams object
// with the ability to set a context for a request.
func NewListGKEClusterSizesParamsWithContext(ctx context.Context) *ListGKEClusterSizesParams {
	return &ListGKEClusterSizesParams{
		Context: ctx,
	}
}

// NewListGKEClusterSizesParamsWithHTTPClient creates a new ListGKEClusterSizesParams object
// with the ability to set a custom HTTPClient for a request.
func NewListGKEClusterSizesParamsWithHTTPClient(client *http.Client) *ListGKEClusterSizesParams {
	return &ListGKEClusterSizesParams{
		HTTPClient: client,
	}
}

/* ListGKEClusterSizesParams contains all the parameters to send to the API endpoint
   for the list g k e cluster sizes operation.

   Typically these are written to a http.Request.
*/
type ListGKEClusterSizesParams struct {

	// ClusterID.
	ClusterID string

	// ProjectID.
	ProjectID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list g k e cluster sizes params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListGKEClusterSizesParams) WithDefaults() *ListGKEClusterSizesParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list g k e cluster sizes params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListGKEClusterSizesParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) WithTimeout(timeout time.Duration) *ListGKEClusterSizesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) WithContext(ctx context.Context) *ListGKEClusterSizesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) WithHTTPClient(client *http.Client) *ListGKEClusterSizesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithClusterID adds the clusterID to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) WithClusterID(clusterID string) *ListGKEClusterSizesParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WithProjectID adds the projectID to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) WithProjectID(projectID string) *ListGKEClusterSizesParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the list g k e cluster sizes params
func (o *ListGKEClusterSizesParams) SetProjectID(projectID string) {
	o.ProjectID = projectID
}

// WriteToRequest writes these params to a swagger request
func (o *ListGKEClusterSizesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	// path param project_id
	if err := r.SetPathParam("project_id", o.ProjectID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
