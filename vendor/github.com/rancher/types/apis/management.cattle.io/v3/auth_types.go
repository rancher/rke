package v3

import (
	extv1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProjectSpec `json:"spec,omitempty"`
}

type ProjectSpec struct {
	DisplayName string `json:"displayName,omitempty" norman:"required"`
	ClusterName string `json:"clusterName,omitempty" norman:"required,type=reference[cluster]"`
}

type ProjectRoleTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Rules   []rbacv1.PolicyRule `json:"rules,omitempty"`
	Builtin bool                `json:"builtin"`

	ProjectRoleTemplateNames []string `json:"projectRoleTemplateNames,omitempty" norman:"type=array[reference[projectRoleTemplate]]"`
}

type PodSecurityPolicyTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec extv1.PodSecurityPolicySpec `json:"spec,omitempty"`
}

type ProjectRoleTemplateBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Subject rbacv1.Subject `json:"subject,omitempty"`

	ProjectName             string `json:"projectName,omitempty" norman:"type=reference[project]"`
	ProjectRoleTemplateName string `json:"projectRoleTemplateName,omitempty" norman:"type=reference[projectRoleTemplate]"`
}

type ClusterRoleTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Rules   []rbacv1.PolicyRule `json:"rules,omitempty"`
	Builtin bool                `json:"builtin"`

	ClusterRoleTemplateNames []string `json:"clusterRoleTemplateNames,omitempty" norman:"type=array[reference[clusterRoleTemplate]]"`
}

type ClusterRoleTemplateBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Subject rbacv1.Subject `json:"subject,omitempty"`

	ClusterName             string `json:"clusterName,omitempty" norman:"type=reference[cluster]"`
	ClusterRoleTemplateName string `json:"clusterRoleTemplateName,omitempty" norman:"type=reference[clusterRoleTemplate]"`
}
