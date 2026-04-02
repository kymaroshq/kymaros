package sandbox

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplyDenyAll creates a deny-all NetworkPolicy that blocks all ingress and egress traffic
func (m *Manager) ApplyDenyAll(ctx context.Context, namespace string) error {
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deny-all",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{}, // selects all pods
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			// Empty ingress/egress = deny all
		},
	}

	if err := m.client.Create(ctx, np); err != nil {
		return fmt.Errorf("create deny-all NetworkPolicy in %q: %w", namespace, err)
	}
	m.logger.InfoContext(ctx, "deny-all NetworkPolicy applied", "namespace", namespace)
	return nil
}

// ApplyGroupAllow creates a NetworkPolicy that allows intra-group traffic
// while still denying traffic from outside the group
func (m *Manager) ApplyGroupAllow(ctx context.Context, namespace string, groupLabel string) error {
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-group",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{}, // selects all pods
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									LabelGroup: groupLabel,
								},
							},
						},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									LabelGroup: groupLabel,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := m.client.Create(ctx, np); err != nil {
		return fmt.Errorf("create allow-group NetworkPolicy in %q: %w", namespace, err)
	}
	m.logger.InfoContext(ctx, "allow-group NetworkPolicy applied", "namespace", namespace, "group", groupLabel)
	return nil
}
