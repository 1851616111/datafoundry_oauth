package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

type User struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`

	// FullName is the full name of user
	FullName string `json:"fullName,omitempty" description:"full name of user"`

	// Identities are the identities associated with this user
	Identities []string `json:"identities" description:"list of identities"`

	// Groups are the groups that this user is a member of
	Groups []string `json:"groups" description:"list of groups"`
}
