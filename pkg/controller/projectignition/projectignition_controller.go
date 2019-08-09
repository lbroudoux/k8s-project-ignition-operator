package projectignition

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	strings "strings"

	yaml "github.com/ghodss/yaml"
	lbroudouxv1beta1 "github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1"
	quotav1 "github.com/openshift/api/quota/v1"
	crqcliv1 "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
	"github.com/redhat-cop/operator-utils/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const controllerName = "controller_projectignition"

var log = logf.Log.WithName("controller_projectignition")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ProjectIgnition Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	//return &ReconcileProjectIgnition{client: mgr.GetClient(), scheme: mgr.GetScheme()}

	// Applying https://github.com/redhat-cop/operator-utils
	return &ReconcileProjectIgnition{
		client:         mgr.GetClient(),
		scheme:         mgr.GetScheme(),
		restConfig:     mgr.GetConfig(),
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetRecorder(controllerName)),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("projectignition-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ProjectIgnition
	err = c.Watch(&source.Kind{Type: &lbroudouxv1beta1.ProjectIgnition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ProjectIgnition
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lbroudouxv1beta1.ProjectIgnition{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileProjectIgnition implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileProjectIgnition{}

// ReconcileProjectIgnition reconciles a ProjectIgnition object
type ReconcileProjectIgnition struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	restConfig *rest.Config
	util.ReconcilerBase
}

// Reconcile reads that state of the cluster for a ProjectIgnition object and makes changes based on the state read
// and what is in the ProjectIgnition.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileProjectIgnition) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ProjectIgnition")

	// Fetch the ProjectIgnition instance
	instance := &lbroudouxv1beta1.ProjectIgnition{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// First validate instance as a ValidatingAdmissionConfiguration is not deployed or available.
	if ok, err := r.IsValid(instance); !ok {
		return r.ManageError(instance, err)
	}

	// Then check if instance is fully initialized.
	if ok := r.IsInitialized(instance); !ok {
		err := r.GetClient().Update(context.TODO(), instance)
		if err != nil {
			log.Error(err, "Unable to update instance", "instance", instance)
			return r.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	// Implemement cleanup logic if needed.
	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, controllerName) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(instance)
		if err != nil {
			log.Error(err, "Unable to delete instance", "instance", instance)
			return r.ManageError(instance, err)
		}
		util.RemoveFinalizer(instance, controllerName)
		err = r.GetClient().Update(context.TODO(), instance)
		if err != nil {
			log.Error(err, "Unable to update instance", "instance", instance)
			return r.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	// Manage Operator logic.
	somethingToDo, err := r.manageOperatorLogic(instance)
	if somethingToDo {
		if err != nil {
			return r.ManageError(instance, err)
		}
		return r.ManageSuccess(instance)
	}
	return reconcile.Result{}, nil
}

// IsValid returns true if CR is semantically correct
func (r *ReconcileProjectIgnition) IsValid(obj metav1.Object) (bool, error) {
	cr, ok := obj.(*lbroudouxv1beta1.ProjectIgnition)
	// validation logic
	if !ok {
		return false, errors.New("Not a ProjectIgnition object")
	}
	if len(cr.Spec.Namespaces.Definitions) > 0 {
		return true, nil
	}
	return false, errors.New("Not valid because no namespaces definitions")
}

func (r *ReconcileProjectIgnition) manageCleanUpLogic(cr *lbroudouxv1beta1.ProjectIgnition) error {
	return nil
}

func (r *ReconcileProjectIgnition) manageOperatorLogic(cr *lbroudouxv1beta1.ProjectIgnition) (bool, error) {
	// First keep trace of changes and check if we're on OpenShift of Vanilla Kubernetes.
	isOpenShift := false
	somethingToDo := false

	// The discovery package is used to discover APIs supported by a Kubernetes API server.
	dclient, err := r.GetDiscoveryClient()
	if err == nil && dclient != nil {
		apiGroupList, err := dclient.ServerGroups()
		if err != nil {
			log.Info("Error while querying ServerGroups, assuming we're on Vanilla Kubernetes")
		} else {
			for i := 0; i < len(apiGroupList.Groups); i++ {
				if strings.HasSuffix(apiGroupList.Groups[i].Name, ".openshift.io") {
					isOpenShift = true
					log.Info("We detected being on OpenShift! Wouhou!")
					break
				}
			}
		}
	} else {
		log.Info("Cannot retrieve a DiscoveryClient, assuming we're on Vanilla Kubernetes")
	}

	// Check each and every namespace definition.
	for i := 0; i < len(cr.Spec.Namespaces.Definitions); i++ {
		var namespaceName = (cr.Spec.ProjectName + "-" + cr.Spec.Namespaces.Definitions[i].Name)

		// Check namespace already exist using a base v1 client.
		v1client, err := kubernetes.NewForConfig(r.restConfig)
		if err != nil {
			panic(err.Error())
		}
		namespace, err := v1client.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			log.Info("Namespace " + namespaceName + " does not exists, creating it")
			somethingToDo = true
			namespace = createNewNamespaceFromDefinition(&cr.Spec, &cr.Spec.Namespaces.Definitions[i])
			namespace, err = v1client.CoreV1().Namespaces().Create(namespace)
			if err != nil {
				log.Error(err, "Failed to create a Namespace")
				return somethingToDo, err
			}
			// Set ownership of namespace to this CR
			if err := controllerutil.SetControllerReference(cr, namespace, r.scheme); err != nil {
				return somethingToDo, err
			}
			_ = append(cr.Status.Namespaces, namespace.Name)

			// Deal with namespace role bindings.
			if cr.Spec.Namespaces.Definitions[i].RoleBindings != nil {
				log.Info("Now dealing with roles in " + namespaceName)
				for j := 0; j < len(cr.Spec.Namespaces.Definitions[i].RoleBindings); j++ {
					roleBinding := createRoleBindingFromDefinition(&cr.Spec, &cr.Spec.Namespaces.Definitions[i], &cr.Spec.Namespaces.Definitions[i].RoleBindings[j])
					err = r.client.Create(context.TODO(), roleBinding)
					if err != nil {
						log.Error(err, "Failed to create a RoleBinding in "+namespaceName)
						return somethingToDo, err
					}

					// Set ownership of RoleBinding to this CR
					if err := controllerutil.SetControllerReference(cr, roleBinding, r.scheme); err != nil {
						return somethingToDo, err
					}
					_ = append(cr.Status.RoleBindings, roleBinding.Name)
				}
			}

			// Deal with namespace resource quotas.
			if cr.Spec.Namespaces.Definitions[i].Quotas != nil {
				log.Info("Now dealing with quotas in " + namespaceName)
				for j := 0; j < len(cr.Spec.Namespaces.Definitions[i].Quotas); j++ {
					quotas, err := createQuotasFromDefinition(&cr.Spec, &cr.Spec.Namespaces.Definitions[i], cr.Spec.Namespaces.Definitions[i].Quotas[j])
					if err != nil {
						log.Error(err, "Failed to download quota file "+cr.Spec.Namespaces.Definitions[i].Quotas[j])
						return somethingToDo, err
					}

					if quotas != nil {
						for k := 0; k < len(quotas); k++ {
							err = r.client.Create(context.TODO(), *quotas[k])
							if err != nil {
								log.Error(err, "Failed to create a ResourceQuota or LimitRange in "+namespaceName)
								return somethingToDo, err
							}
						}
					}
				}
			}
		} else {
			log.Info("Namespace " + namespace.Name + " already exists")
		}
	}

	// Now deal with globale cluster resource quota if we're on OpenShift.
	if cr.Spec.OpenShiftMultiProjectQuota.Quota != "" && isOpenShift {
		crqclient, err := crqcliv1.NewForConfig(r.restConfig)
		if err != nil {
			log.Error(err, "Failed to retrieve a ClusterResourceQuota client")
			return somethingToDo, err
		}

		clusterquota, err := crqclient.ClusterResourceQuotas().Get(cr.Spec.ProjectName+"-quota", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			log.Info("ClusterResourceQuota does not exists, creating it")
			somethingToDo = true
			clusterquota, err = createClusterResourceQuotaFromSpec(&cr.Spec)
			if err != nil {
				log.Error(err, "Failed to prepare ClusterResourceQuota from file "+cr.Spec.OpenShiftMultiProjectQuota.Quota)
				return somethingToDo, err
			}

			if clusterquota != nil {
				clusterquota, err = crqclient.ClusterResourceQuotas().Create(clusterquota)
				if err != nil {
					log.Error(err, "Failed to create a ClusterResourceQuota")
					return somethingToDo, err
				}
			}
		}
	}

	return somethingToDo, nil
}

func createNewNamespaceFromDefinition(spec *lbroudouxv1beta1.ProjectIgnitionSpec, def *lbroudouxv1beta1.DefinitionSpec) *corev1.Namespace {
	namespaceName := spec.ProjectName + "-" + def.Name
	// Create annotations if specified.
	annotations := map[string]string{}
	if def.Annotations != nil {
		for i := 0; i < len(def.Annotations); i++ {
			split := strings.Split(def.Annotations[i], ": ")
			annotations[split[0]] = replaceProjectOrNamespaceInString(spec.ProjectName, namespaceName, split[1])
		}
	}
	// Create labels if specified.
	labels := map[string]string{}
	if def.Labels != nil {
		for i := 0; i < len(def.Labels); i++ {
			labels[def.Labels[i].Key] = def.Labels[i].Value
		}
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespaceName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.NamespaceSpec{},
	}
}

func createRoleBindingFromDefinition(spec *lbroudouxv1beta1.ProjectIgnitionSpec, def *lbroudouxv1beta1.DefinitionSpec, roleSpec *lbroudouxv1beta1.RoleBindingSpec) *rbacv1.RoleBinding {
	namespaceName := spec.ProjectName + "-" + def.Name
	if roleSpec.User != "" {
		return &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleSpec.Role + "_" + replaceProjectOrNamespaceInString(spec.ProjectName, namespaceName, roleSpec.User),
				Namespace: namespaceName,
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "Role",
				Name:     roleSpec.Role,
				APIGroup: "rbac.authorization.k8s.io",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     "User",
					Name:     replaceProjectOrNamespaceInString(spec.ProjectName, namespaceName, roleSpec.User),
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		}
	} else {
		return &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleSpec.Role + "_" + replaceProjectOrNamespaceInString(spec.ProjectName, namespaceName, roleSpec.Group),
				Namespace: namespaceName,
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "Role",
				Name:     roleSpec.Role,
				APIGroup: "rbac.authorization.k8s.io",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     "Group",
					Name:     replaceProjectOrNamespaceInString(spec.ProjectName, namespaceName, roleSpec.Group),
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		}
	}
}

func createQuotasFromDefinition(spec *lbroudouxv1beta1.ProjectIgnitionSpec, def *lbroudouxv1beta1.DefinitionSpec, quotaURL string) ([]*runtime.Object, error) {
	namespaceName := spec.ProjectName + "-" + def.Name
	if strings.HasPrefix(quotaURL, "http://") || strings.HasPrefix(quotaURL, "https:/") {
		bytes, err := downloadFile(quotaURL)
		if err != nil {
			return nil, err
		}

		acceptedK8sTypes := regexp.MustCompile(`(ResourceQuota|LimitRange)`)
		fileAsString := string(bytes[:])
		sepYamlfiles := strings.Split(fileAsString, "---")
		retVal := make([]*runtime.Object, 0, len(sepYamlfiles))
		for _, f := range sepYamlfiles {
			if f == "\n" || f == "" {
				// ignore empty cases
				continue
			}

			decode := scheme.Codecs.UniversalDeserializer().Decode
			obj, groupVersionKind, err := decode([]byte(f), nil, nil)
			if err != nil {
				log.Error(err, "Error while decoding YAML object")
				continue
			}

			if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
				log.Info("The quota file contained K8s object types which are not supported! Skipping object with type: " + groupVersionKind.Kind)
			} else {
				switch o := obj.(type) {
				case *corev1.ResourceQuota:
					rq := obj.(*corev1.ResourceQuota)
					rq.Namespace = namespaceName
					break
				case *corev1.LimitRange:
					lr := obj.(*corev1.LimitRange)
					lr.Namespace = namespaceName
					break
				default:
					_ = o
				}
				retVal = append(retVal, &obj)
			}
		}
		return retVal, nil
	}
	return nil, nil
}

func createClusterResourceQuotaFromSpec(spec *lbroudouxv1beta1.ProjectIgnitionSpec) (*quotav1.ClusterResourceQuota, error) {
	bytes, err := downloadFile(spec.OpenShiftMultiProjectQuota.Quota)
	if err != nil {
		log.Error(err, "Error while downloading ClusterResourceQuota template")
		return nil, err
	}

	var crq quotav1.ClusterResourceQuota
	if strings.HasSuffix(spec.OpenShiftMultiProjectQuota.Quota, ".yaml") || strings.HasSuffix(spec.OpenShiftMultiProjectQuota.Quota, ".yml") {
		err = yaml.Unmarshal(bytes, &crq)
	} else {
		err = json.Unmarshal(bytes, &crq)
	}
	if err != nil {
		log.Error(err, "Error while unmarshalling YAML or JSON to ClusterResourceQuota struct")
		return nil, err
	}

	// Now set crq properties.
	crq.SetName(spec.ProjectName + "-quota")

	// Create annotations selector if specified.
	if spec.OpenShiftMultiProjectQuota.ProjectAnnotationSelector != "" {
		annotations := map[string]string{}
		split := strings.Split(spec.OpenShiftMultiProjectQuota.ProjectAnnotationSelector, ": ")
		annotations[split[0]] = replaceProjectOrNamespaceInString(spec.ProjectName, "", split[1])
		crq.Spec.Selector.AnnotationSelector = annotations
	}
	// Create label seclector if specified.
	if spec.OpenShiftMultiProjectQuota.ProjectLabelSelector != "" {
		labels := map[string]string{}
		split := strings.Split(spec.OpenShiftMultiProjectQuota.ProjectLabelSelector, ": ")
		labels[split[0]] = replaceProjectOrNamespaceInString(spec.ProjectName, "", split[1])
		crq.Spec.Selector.LabelSelector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
	}
	return &crq, nil
}

func replaceProjectOrNamespaceInString(projectName string, namespaceName string, stringToReplace string) string {
	if strings.Contains(stringToReplace, "{project}") {
		stringToReplace = strings.Replace(stringToReplace, "{project}", projectName, -1)
	}
	if strings.Contains(stringToReplace, "{namespace}") {
		stringToReplace = strings.Replace(stringToReplace, "{namespace}", namespaceName, -1)
	}
	return stringToReplace
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
