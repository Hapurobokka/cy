package layout

import (
	"context"
	"reflect"
	"testing"

	"github.com/cfoust/cy/pkg/janet"

	"github.com/stretchr/testify/require"
)

func TestAttachFirst(t *testing.T) {
	require.Equal(t, &TabsNode{
		Tabs: []Tab{
			{
				Name:   "foo",
				Active: true,
				Node: &ViewNode{
					Attached: true,
				},
			},
		},
	}, AttachFirst(&TabsNode{
		Tabs: []Tab{
			{
				Name:   "foo",
				Active: true,
				Node: &ViewNode{
					Attached: false,
				},
			},
		},
	}))

	require.Equal(t,
		&SplitNode{
			A: &MarginsNode{Node: &ViewNode{Attached: true}},
			B: &ViewNode{},
		},
		AttachFirst(&SplitNode{
			A: &MarginsNode{Node: &ViewNode{}},
			B: &ViewNode{},
		}),
	)
}

func TestAttachFirstStack(t *testing.T) {
	require.Equal(t, &StackNode{
		Leaves: []Leaf{
			{
				Active: true,
				Node: &ViewNode{
					Attached: true,
				},
			},
		},
	}, AttachFirst(&StackNode{
		Leaves: []Leaf{
			{
				Active: true,
				Node: &ViewNode{
					Attached: false,
				},
			},
		},
	}))
}

func TestRemoveAttached(t *testing.T) {
	require.Equal(t,
		&MarginsNode{Node: &ViewNode{Attached: true}},
		RemoveAttached(&SplitNode{
			A: &MarginsNode{Node: &ViewNode{}},
			B: &ViewNode{Attached: true},
		}),
	)
}

func TestRemoveAttachedStack(t *testing.T) {
	// With two leaves, removing the attached one should return
	// a stack with the remaining leaf (now active and attached)
	require.Equal(t,
		&StackNode{
			Leaves: []Leaf{
				{
					Active: true,
					Node:   &ViewNode{Attached: true},
				},
			},
		},
		RemoveAttached(&StackNode{
			Leaves: []Leaf{
				{
					Active: true,
					Node:   &ViewNode{Attached: true},
				},
				{
					Node: &ViewNode{},
				},
			},
		}),
	)

	// With one leaf, removing should unwrap (preserving attached state)
	require.Equal(t,
		&ViewNode{Attached: true},
		RemoveAttached(&StackNode{
			Leaves: []Leaf{
				{
					Active: true,
					Node:   &ViewNode{Attached: true},
				},
			},
		}),
	)
}

func TestAttach(t *testing.T) {
	var id int32 = 1
	node := Attach(&SplitNode{
		A: &ViewNode{Attached: false},
		B: &ViewNode{Attached: true},
	}, id)

	require.Equal(t,
		id,
		*node.Children()[1].(*ViewNode).ID,
	)
}

func TestTabsHideBar(t *testing.T) {
	ctx := context.Background()
	vm, err := janet.New(ctx)
	require.NoError(t, err)

	unmarshal := func(t *testing.T, code string) *TabsNode {
		result, err := vm.ExecuteCall(ctx, nil, janet.CallString(code))
		require.NoError(t, err)
		require.NotNil(t, result.Yield)

		node, err := (&TabsNode{}).UnmarshalJanet(result.Yield)
		require.NoError(t, err)

		tabs, ok := node.(*TabsNode)
		require.True(t, ok)
		return tabs
	}

	t.Run("set", func(t *testing.T) {
		tabs := unmarshal(t, `(yield
			{:type :tabs
			 :hide-bar true
			 :tabs @[{:active true :name "pane" :node {:type :view :attached true}}]})`)
		require.True(t, tabs.HideBar)

		// The flag is also emitted when marshalling back to Janet
		marshaled := reflect.ValueOf(tabs.MarshalJanet())
		require.True(t, marshaled.FieldByName("HideBar").Bool())
	})

	t.Run("default", func(t *testing.T) {
		tabs := unmarshal(t, `(yield
			{:type :tabs
			 :tabs @[{:active true :name "pane" :node {:type :view :attached true}}]})`)
		require.False(t, tabs.HideBar)
	})
}
