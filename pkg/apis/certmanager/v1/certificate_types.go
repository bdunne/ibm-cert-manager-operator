//
// Copyright 2022 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/ibm/ibm-cert-manager-operator/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A Certificate resource should be created to ensure an up to date and signed
// x509 certificate is stored in the Kubernetes Secret resource named in `spec.secretName`.
//
// The stored certificate will be renewed before it expires (as configured by `spec.renewBefore`).
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certificates,scope=Namespaced
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".spec.secretName",description=""
// +kubebuilder:printcolumn:name="Issuer",type="string",JSONPath=".spec.issuerRef.name",description="",priority=1
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:printcolumn:name="Expiration",type="string",JSONPath=".status.notAfter"
// +kubebuilder:resource:path=certificates,shortName=cert;certs,categories=cert-manager,scope=Namespaced
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Desired state of the Certificate resource.
	Spec CertificateSpec `json:"spec"`

	// Status of the Certificate. This is set and managed automatically.
	// +optional
	Status CertificateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateList contains a list of Certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

// +kubebuilder:validation:Enum=RSA;ECDSA
type PrivateKeyAlgorithm string

const (
	// Denotes the RSA private key type.
	RSAKeyAlgorithm PrivateKeyAlgorithm = "RSA"

	// Denotes the ECDSA private key type.
	ECDSAKeyAlgorithm PrivateKeyAlgorithm = "ECDSA"
)

// +kubebuilder:validation:Enum=PKCS1;PKCS8
type PrivateKeyEncoding string

const (
	// PKCS1 key encoding will produce PEM files that include the type of
	// private key as part of the PEM header, e.g. `BEGIN RSA PRIVATE KEY`.
	// If the keyAlgorithm is set to 'ECDSA', this will produce private keys
	// that use the `BEGIN EC PRIVATE KEY` header.
	PKCS1 PrivateKeyEncoding = "PKCS1"

	// PKCS8 key encoding will produce PEM files with the `BEGIN PRIVATE KEY`
	// header. It encodes the keyAlgorithm of the private key as part of the
	// DER encoded PEM block.
	PKCS8 PrivateKeyEncoding = "PKCS8"
)

// CertificateSpec defines the desired state of Certificate.
// A valid Certificate requires at least one of a CommonName, DNSName, or
// URISAN to be valid.
type CertificateSpec struct {
	// Full X509 name specification (https://golang.org/pkg/crypto/x509/pkix/#Name).
	// +optional
	Subject *X509Subject `json:"subject,omitempty"`

	// CommonName is a common name to be used on the Certificate.
	// The CommonName should have a length of 64 characters or fewer to avoid
	// generating invalid CSRs.
	// This value is ignored by TLS clients when any subject alt name is set.
	// This is x509 behaviour: https://tools.ietf.org/html/rfc6125#section-6.4.4
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// The requested 'duration' (i.e. lifetime) of the Certificate. This option
	// may be ignored/overridden by some issuer types. If unset this defaults to
	// 90 days. Certificate will be renewed either 2/3 through its duration or
	// `renewBefore` period before its expiry, whichever is later. Minimum
	// accepted duration is 1 hour. Value must be in units accepted by Go
	// time.ParseDuration https://golang.org/pkg/time/#ParseDuration
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// How long before the currently issued certificate's expiry
	// cert-manager should renew the certificate. The default is 2/3 of the
	// issued certificate's duration. Minimum accepted value is 5 minutes.
	// Value must be in units accepted by Go time.ParseDuration
	// https://golang.org/pkg/time/#ParseDuration
	// +optional
	RenewBefore *metav1.Duration `json:"renewBefore,omitempty"`

	// DNSNames is a list of DNS subjectAltNames to be set on the Certificate.
	// +optional
	DNSNames []string `json:"dnsNames,omitempty"`

	// IPAddresses is a list of IP address subjectAltNames to be set on the Certificate.
	// +optional
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// URIs is a list of URI subjectAltNames to be set on the Certificate.
	// +optional
	URIs []string `json:"uris,omitempty"`

	// EmailAddresses is a list of email subjectAltNames to be set on the Certificate.
	// +optional
	EmailAddresses []string `json:"emailAddresses,omitempty"`

	// SecretName is the name of the secret resource that will be automatically
	// created and managed by this Certificate resource.
	// It will be populated with a private key and certificate, signed by the
	// denoted issuer.
	SecretName string `json:"secretName"`

	// Keystores configures additional keystore output formats stored in the
	// `secretName` Secret resource.
	// +optional
	Keystores *CertificateKeystores `json:"keystores,omitempty"`

	// IssuerRef is a reference to the issuer for this certificate.
	// If the `kind` field is not set, or set to `Issuer`, an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the `kind` field is set to `ClusterIssuer`, a ClusterIssuer with the
	// provided name will be used.
	// The `name` field in this stanza is required at all times.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`

	// IsCA will mark this Certificate as valid for certificate signing.
	// This will automatically add the `cert sign` usage to the list of `usages`.
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Usages is the set of x509 usages that are requested for the certificate.
	// Defaults to `digital signature` and `key encipherment` if not specified.
	// +optional
	Usages []KeyUsage `json:"usages,omitempty"`

	// Options to control private keys used for the Certificate.
	// +optional
	PrivateKey *CertificatePrivateKey `json:"privateKey,omitempty"`

	// EncodeUsagesInRequest controls whether key usages should be present
	// in the CertificateRequest
	// +optional
	EncodeUsagesInRequest *bool `json:"encodeUsagesInRequest,omitempty"`

	// revisionHistoryLimit is the maximum number of CertificateRequest revisions
	// that are maintained in the Certificate's history. Each revision represents
	// a single `CertificateRequest` created by this Certificate, either when it
	// was created, renewed, or Spec was changed. Revisions will be removed by
	// oldest first if the number of revisions exceeds this number. If set,
	// revisionHistoryLimit must be a value of `1` or greater. If unset (`nil`),
	// revisions will not be garbage collected. Default value is `nil`.
	// +kubebuilder:validation:ExclusiveMaximum=false
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"` // Validated by the validating webhook.
}

// CertificatePrivateKey contains configuration options for private keys
// used by the Certificate controller.
// This allows control of how private keys are rotated.
type CertificatePrivateKey struct {
	// RotationPolicy controls how private keys should be regenerated when a
	// re-issuance is being processed.
	// If set to Never, a private key will only be generated if one does not
	// already exist in the target `spec.secretName`. If one does exists but it
	// does not have the correct algorithm or size, a warning will be raised
	// to await user intervention.
	// If set to Always, a private key matching the specified requirements
	// will be generated whenever a re-issuance occurs.
	// Default is 'Never' for backward compatibility.
	// +optional
	RotationPolicy PrivateKeyRotationPolicy `json:"rotationPolicy,omitempty"`

	// The private key cryptography standards (PKCS) encoding for this
	// certificate's private key to be encoded in.
	// If provided, allowed values are `PKCS1` and `PKCS8` standing for PKCS#1
	// and PKCS#8, respectively.
	// Defaults to `PKCS1` if not specified.
	// +optional
	Encoding PrivateKeyEncoding `json:"encoding,omitempty"`

	// Algorithm is the private key algorithm of the corresponding private key
	// for this certificate. If provided, allowed values are either `RSA` or `ECDSA`
	// If `algorithm` is specified and `size` is not provided,
	// key size of 256 will be used for `ECDSA` key algorithm and
	// key size of 2048 will be used for `RSA` key algorithm.
	// +optional
	Algorithm PrivateKeyAlgorithm `json:"algorithm,omitempty"`

	// Size is the key bit size of the corresponding private key for this certificate.
	// If `algorithm` is set to `RSA`, valid values are `2048`, `4096` or `8192`,
	// and will default to `2048` if not specified.
	// If `algorithm` is set to `ECDSA`, valid values are `256`, `384` or `521`,
	// and will default to `256` if not specified.
	// No other values are allowed.
	// +optional
	Size int `json:"size,omitempty"` // Validated by webhook. Be mindful of adding OpenAPI validation- see https://github.com/jetstack/cert-manager/issues/3644
}

// Denotes how private keys should be generated or sourced when a Certificate
// is being issued.
type PrivateKeyRotationPolicy string

var (
	// RotationPolicyNever means a private key will only be generated if one
	// does not already exist in the target `spec.secretName`.
	// If one does exists but it does not have the correct algorithm or size,
	// a warning will be raised to await user intervention.
	RotationPolicyNever PrivateKeyRotationPolicy = "Never"

	// RotationPolicyAlways means a private key matching the specified
	// requirements will be generated whenever a re-issuance occurs.
	RotationPolicyAlways PrivateKeyRotationPolicy = "Always"
)

// X509Subject Full X509 name specification
type X509Subject struct {
	// Organizations to be used on the Certificate.
	// +optional
	Organizations []string `json:"organizations,omitempty"`
	// Countries to be used on the Certificate.
	// +optional
	Countries []string `json:"countries,omitempty"`
	// Organizational Units to be used on the Certificate.
	// +optional
	OrganizationalUnits []string `json:"organizationalUnits,omitempty"`
	// Cities to be used on the Certificate.
	// +optional
	Localities []string `json:"localities,omitempty"`
	// State/Provinces to be used on the Certificate.
	// +optional
	Provinces []string `json:"provinces,omitempty"`
	// Street addresses to be used on the Certificate.
	// +optional
	StreetAddresses []string `json:"streetAddresses,omitempty"`
	// Postal codes to be used on the Certificate.
	// +optional
	PostalCodes []string `json:"postalCodes,omitempty"`
	// Serial number to be used on the Certificate.
	// +optional
	SerialNumber string `json:"serialNumber,omitempty"`
}

// CertificateKeystores configures additional keystore output formats to be
// created in the Certificate's output Secret.
type CertificateKeystores struct {
	// JKS configures options for storing a JKS keystore in the
	// `spec.secretName` Secret resource.
	// +optional
	JKS *JKSKeystore `json:"jks,omitempty"`

	// PKCS12 configures options for storing a PKCS12 keystore in the
	// `spec.secretName` Secret resource.
	// +optional
	PKCS12 *PKCS12Keystore `json:"pkcs12,omitempty"`
}

// JKS configures options for storing a JKS keystore in the `spec.secretName`
// Secret resource.
type JKSKeystore struct {
	// Create enables JKS keystore creation for the Certificate.
	// If true, a file named `keystore.jks` will be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef`.
	// The keystore file will only be updated upon re-issuance.
	// A file named `truststore.jks` will also be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef` containing the issuing Certificate Authority
	Create bool `json:"create"`

	// PasswordSecretRef is a reference to a key in a Secret resource
	// containing the password used to encrypt the JKS keystore.
	PasswordSecretRef cmmeta.SecretKeySelector `json:"passwordSecretRef"`
}

// PKCS12 configures options for storing a PKCS12 keystore in the
// `spec.secretName` Secret resource.
type PKCS12Keystore struct {
	// Create enables PKCS12 keystore creation for the Certificate.
	// If true, a file named `keystore.p12` will be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef`.
	// The keystore file will only be updated upon re-issuance.
	// A file named `truststore.p12` will also be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef` containing the issuing Certificate Authority
	Create bool `json:"create"`

	// PasswordSecretRef is a reference to a key in a Secret resource
	// containing the password used to encrypt the PKCS12 keystore.
	PasswordSecretRef cmmeta.SecretKeySelector `json:"passwordSecretRef"`
}

// CertificateStatus defines the observed state of Certificate
type CertificateStatus struct {
	// List of status conditions to indicate the status of certificates.
	// Known condition types are `Ready` and `Issuing`.
	// +optional
	Conditions []CertificateCondition `json:"conditions,omitempty"`

	// LastFailureTime is the time as recorded by the Certificate controller
	// of the most recent failure to complete a CertificateRequest for this
	// Certificate resource.
	// If set, cert-manager will not re-request another Certificate until
	// 1 hour has elapsed from this time.
	// +optional
	LastFailureTime *metav1.Time `json:"lastFailureTime,omitempty"`

	// The time after which the certificate stored in the secret named
	// by this resource in spec.secretName is valid.
	// +optional
	NotBefore *metav1.Time `json:"notBefore,omitempty"`

	// The expiration time of the certificate stored in the secret named
	// by this resource in `spec.secretName`.
	// +optional
	NotAfter *metav1.Time `json:"notAfter,omitempty"`

	// RenewalTime is the time at which the certificate will be next
	// renewed.
	// If not set, no upcoming renewal is scheduled.
	// +optional
	RenewalTime *metav1.Time `json:"renewalTime,omitempty"`

	// The current 'revision' of the certificate as issued.
	//
	// When a CertificateRequest resource is created, it will have the
	// `cert-manager.io/certificate-revision` set to one greater than the
	// current value of this field.
	//
	// Upon issuance, this field will be set to the value of the annotation
	// on the CertificateRequest resource used to issue the certificate.
	//
	// Persisting the value on the CertificateRequest resource allows the
	// certificates controller to know whether a request is part of an old
	// issuance or if it is part of the ongoing revision's issuance by
	// checking if the revision value in the annotation is greater than this
	// field.
	// +optional
	Revision *int `json:"revision,omitempty"`

	// The name of the Secret resource containing the private key to be used
	// for the next certificate iteration.
	// The keymanager controller will automatically set this field if the
	// `Issuing` condition is set to `True`.
	// It will automatically unset this field when the Issuing condition is
	// not set or False.
	// +optional
	NextPrivateKeySecretName *string `json:"nextPrivateKeySecretName,omitempty"`
}

// CertificateCondition contains condition information for an Certificate.
type CertificateCondition struct {
	// Type of the condition, known values are (`Ready`, `Issuing`).
	Type CertificateConditionType `json:"type"`

	// Status of the condition, one of (`True`, `False`, `Unknown`).
	Status cmmeta.ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`

	// If set, this represents the .metadata.generation that the condition was
	// set based upon.
	// For instance, if .metadata.generation is currently 12, but the
	// .status.condition[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the Certificate.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CertificateConditionType represents an Certificate condition value.
type CertificateConditionType string

const (
	// CertificateConditionReady indicates that a certificate is ready for use.
	// This is defined as:
	// - The target secret exists
	// - The target secret contains a certificate that has not expired
	// - The target secret contains a private key valid for the certificate
	// - The commonName and dnsNames attributes match those specified on the Certificate
	CertificateConditionReady CertificateConditionType = "Ready"

	// A condition added to Certificate resources when an issuance is required.
	// This condition will be automatically added and set to true if:
	//   * No keypair data exists in the target Secret
	//   * The data stored in the Secret cannot be decoded
	//   * The private key and certificate do not have matching public keys
	//   * If a CertificateRequest for the current revision exists and the
	//     certificate data stored in the Secret does not match the
	//    `status.certificate` on the CertificateRequest.
	//   * If no CertificateRequest resource exists for the current revision,
	//     the options on the Certificate resource are compared against the
	//     x509 data in the Secret, similar to what's done in earlier versions.
	//     If there is a mismatch, an issuance is triggered.
	// This condition may also be added by external API consumers to trigger
	// a re-issuance manually for any other reason.
	//
	// It will be removed by the 'issuing' controller upon completing issuance.
	CertificateConditionIssuing CertificateConditionType = "Issuing"
)

func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
