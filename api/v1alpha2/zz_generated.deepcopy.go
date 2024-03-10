//go:build !ignore_autogenerated

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Aggregate) DeepCopyInto(out *Aggregate) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Aggregate.
func (in *Aggregate) DeepCopy() *Aggregate {
	if in == nil {
		return nil
	}
	out := new(Aggregate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IscsiSubSpec) DeepCopyInto(out *IscsiSubSpec) {
	*out = *in
	if in.Lifs != nil {
		in, out := &in.Lifs, &out.Lifs
		*out = make([]LIF, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IscsiSubSpec.
func (in *IscsiSubSpec) DeepCopy() *IscsiSubSpec {
	if in == nil {
		return nil
	}
	out := new(IscsiSubSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LIF) DeepCopyInto(out *LIF) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LIF.
func (in *LIF) DeepCopy() *LIF {
	if in == nil {
		return nil
	}
	out := new(LIF)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsExport) DeepCopyInto(out *NfsExport) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]NfsRule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsExport.
func (in *NfsExport) DeepCopy() *NfsExport {
	if in == nil {
		return nil
	}
	out := new(NfsExport)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsRule) DeepCopyInto(out *NfsRule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsRule.
func (in *NfsRule) DeepCopy() *NfsRule {
	if in == nil {
		return nil
	}
	out := new(NfsRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsSubSpec) DeepCopyInto(out *NfsSubSpec) {
	*out = *in
	if in.Lifs != nil {
		in, out := &in.Lifs, &out.Lifs
		*out = make([]LIF, len(*in))
		copy(*out, *in)
	}
	if in.Export != nil {
		in, out := &in.Export, &out.Export
		*out = new(NfsExport)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsSubSpec.
func (in *NfsSubSpec) DeepCopy() *NfsSubSpec {
	if in == nil {
		return nil
	}
	out := new(NfsSubSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageVirtualMachine) DeepCopyInto(out *StorageVirtualMachine) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageVirtualMachine.
func (in *StorageVirtualMachine) DeepCopy() *StorageVirtualMachine {
	if in == nil {
		return nil
	}
	out := new(StorageVirtualMachine)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageVirtualMachine) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageVirtualMachineList) DeepCopyInto(out *StorageVirtualMachineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StorageVirtualMachine, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageVirtualMachineList.
func (in *StorageVirtualMachineList) DeepCopy() *StorageVirtualMachineList {
	if in == nil {
		return nil
	}
	out := new(StorageVirtualMachineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageVirtualMachineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageVirtualMachineSpec) DeepCopyInto(out *StorageVirtualMachineSpec) {
	*out = *in
	if in.Aggregates != nil {
		in, out := &in.Aggregates, &out.Aggregates
		*out = make([]Aggregate, len(*in))
		copy(*out, *in)
	}
	if in.ManagementLIF != nil {
		in, out := &in.ManagementLIF, &out.ManagementLIF
		*out = new(LIF)
		**out = **in
	}
	out.ClusterCredentialSecret = in.ClusterCredentialSecret
	out.VsadminCredentialSecret = in.VsadminCredentialSecret
	if in.NfsConfig != nil {
		in, out := &in.NfsConfig, &out.NfsConfig
		*out = new(NfsSubSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.IscsiConfig != nil {
		in, out := &in.IscsiConfig, &out.IscsiConfig
		*out = new(IscsiSubSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageVirtualMachineSpec.
func (in *StorageVirtualMachineSpec) DeepCopy() *StorageVirtualMachineSpec {
	if in == nil {
		return nil
	}
	out := new(StorageVirtualMachineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageVirtualMachineStatus) DeepCopyInto(out *StorageVirtualMachineStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageVirtualMachineStatus.
func (in *StorageVirtualMachineStatus) DeepCopy() *StorageVirtualMachineStatus {
	if in == nil {
		return nil
	}
	out := new(StorageVirtualMachineStatus)
	in.DeepCopyInto(out)
	return out
}
