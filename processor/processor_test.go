package processor

import (
	"testing"

	"github.com/elastic/crd-ref-docs/config"
	"github.com/elastic/crd-ref-docs/types"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func process(t *testing.T, path string) []types.GroupVersionDetails {
	gvd, err := Process(&config.Config{
		Flags: config.Flags{
			SourcePath: "../test/api",
			MaxDepth:   10,
		},
	})
	require.NoError(t, err, "Unable to process")
	return gvd
}

func TestProcessor(t *testing.T) {
	t.Run("basic parsing", func(t *testing.T) {
		gvds := process(t, "../test/api")

		require.Len(t, gvds, 1)
		gvd := gvds[0]

		require.Equal(t, gvd.GroupVersion, schema.GroupVersion{Group: "webapp.test.k8s.elastic.co", Version: "v1"})
		require.Equal(t, gvd.Doc, "Package v1 contains API Schema definitions for the webapp v1 API group\n")
		require.Len(t, gvd.Kinds, 3)
		require.Contains(t, gvd.Kinds, "Embedded")
		require.Contains(t, gvd.Kinds, "Guestbook")
		require.Contains(t, gvd.Kinds, "GuestbookList")

		require.Len(t, gvd.Types, 13)
		require.Contains(t, gvd.Types, "Embedded")
		require.Contains(t, gvd.Types, "Embedded1")
		require.Contains(t, gvd.Types, "Embedded2")
		require.Contains(t, gvd.Types, "Embedded3")
		require.Contains(t, gvd.Types, "Embedded4")
		require.Contains(t, gvd.Types, "EmbeddedX")
		require.Contains(t, gvd.Types, "GuestbookList")
		require.Contains(t, gvd.Types, "Guestbook")
		require.Contains(t, gvd.Types, "GuestbookSpec")
		require.Contains(t, gvd.Types, "GuestbookStatus")
		require.Contains(t, gvd.Types, "GuestbookEntry")
		require.Contains(t, gvd.Types, "GuestbookHeader")
		require.Contains(t, gvd.Types, "Rating")
	})

	t.Run("root object", func(t *testing.T) {
		gvds := process(t, "../test/api")
		guestbook := gvds[0].Types["Guestbook"]

		// Clear sub-fields/references for easier testing
		guestbook.References[0].Fields = nil
		guestbook.Fields[2].Type.Fields = nil
		guestbook.Fields[2].Type.References = nil
		guestbook.Fields[3].Type.Fields = nil
		guestbook.Fields[3].Type.References = nil
		guestbook.Fields[4].Type.Fields = nil
		guestbook.Fields[4].Type.References = nil

		require.Equal(t, guestbook, &types.Type{
			GVK: &schema.GroupVersionKind{
				Group:   "webapp.test.k8s.elastic.co",
				Version: "v1",
				Kind:    "Guestbook",
			},
			Package: "github.com/elastic/crd-ref-docs/api/v1",
			Doc:     "Guestbook is the Schema for the guestbooks API.",
			Name:    "Guestbook",
			Kind:    types.StructKind,
			Fields: []*types.Field{
				{
					Name: "kind",
					Doc:  "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "apiVersion",
					Doc:  "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name:     "metadata",
					Embedded: true,
					Type: &types.Type{
						Name:       "ObjectMeta",
						Package:    "k8s.io/apimachinery/pkg/apis/meta/v1",
						Doc:        "ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.",
						Kind:       types.StructKind,
						Fields:     nil, // Cleared
						References: nil, // Cleared
					},
				},
				{
					Name: "spec",
					Type: &types.Type{
						Name:       "GuestbookSpec",
						Package:    "github.com/elastic/crd-ref-docs/api/v1",
						Doc:        "GuestbookSpec defines the desired state of Guestbook.",
						Kind:       types.StructKind,
						Fields:     nil, // Cleared
						References: nil, // Cleared
					},
				},
				{
					Name: "status",
					Type: &types.Type{
						Name:       "GuestbookStatus",
						Package:    "github.com/elastic/crd-ref-docs/api/v1",
						Doc:        "GuestbookStatus defines the observed state of Guestbook.",
						Kind:       types.StructKind,
						Fields:     nil, // Cleared
						References: nil, // Cleared
					},
				},
			},
			References: []*types.Type{
				{
					Name:    "GuestbookList",
					Package: "github.com/elastic/crd-ref-docs/api/v1",
					Doc:     "GuestbookList contains a list of Guestbook.",
					GVK: &schema.GroupVersionKind{
						Group:   "webapp.test.k8s.elastic.co",
						Version: "v1",
						Kind:    "GuestbookList",
					},
					Kind:   types.StructKind,
					Fields: nil, // Cleared
				},
			},
		})
	})

	t.Run("alias parsing", func(t *testing.T) {
		gvds := process(t, "../test/api")
		rating := gvds[0].Types["Rating"]

		// Clear sub-fields/references for easier testing
		rating.References[0].References = nil
		rating.References[0].Fields = nil

		require.Equal(t, rating, &types.Type{
			Name:    "Rating",
			Package: "github.com/elastic/crd-ref-docs/api/v1",
			Doc:     "Rating is the rating provided by a guest.",
			Kind:    types.AliasKind,
			UnderlyingType: &types.Type{
				Name: "string",
				Kind: types.BasicKind,
			},
			References: []*types.Type{
				{
					Name:       "GuestbookEntry",
					Package:    "github.com/elastic/crd-ref-docs/api/v1",
					Doc:        "GuestbookEntry defines an entry in a guest book.",
					Kind:       types.StructKind,
					References: nil, // Cleared
				},
			},
		})
	})

	t.Run("embedded parsing", func(t *testing.T) {
		gvds := process(t, "../test/api")
		embedded := gvds[0].Types["Embedded"]

		// Clear sub-fields/references for easier testing
		embedded.Fields[2].Type.Fields = nil
		embedded.Fields[2].Type.References = nil

		require.Equal(t, embedded, &types.Type{
			GVK: &schema.GroupVersionKind{
				Group:   "webapp.test.k8s.elastic.co",
				Version: "v1",
				Kind:    "Embedded",
			},
			Package: "github.com/elastic/crd-ref-docs/api/v1",
			Name:    "Embedded",
			Kind:    types.StructKind,
			Fields: []*types.Field{
				{
					Name: "kind",
					Doc:  "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "apiVersion",
					Doc:  "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name:     "metadata",
					Embedded: true,
					Type: &types.Type{
						Name:       "ObjectMeta",
						Package:    "k8s.io/apimachinery/pkg/apis/meta/v1",
						Doc:        "ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.",
						Kind:       types.StructKind,
						Fields:     nil, // Cleared
						References: nil, // Cleared
					},
				},
				{
					Name: "a",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "b",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "c",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "x",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "d",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
				{
					Name: "e",
					Type: &types.Type{
						Name: "string",
						Kind: types.BasicKind,
					},
				},
			},
		})
	})
}
