# easy-olm-operator

> an operator that makes olm easy to use outside of openshift

## features

* create operator group automatically for every namespace
* easily install and manage crds with CrdRef resource (this does ref counting so the crds will be removed when nobody is using them)
* automatically approve the first installplan when in manual mode
