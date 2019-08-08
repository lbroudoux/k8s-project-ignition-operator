package projectignition

import (
	"context"
	"errors"
	strings "strings"

	lbroudouxv1beta1 "github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1"
	"github.com/redhat-cop/operator-utils/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
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
	err = r.manageOperatorLogic(instance)
	if err != nil {
		return r.ManageError(instance, err)
	}
	return r.ManageSuccess(instance)

	/*
		// Set ProjectIgnition instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, namespace, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			err = r.client.Create(context.TODO(), pod)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
		return reconcile.Result{}, nil
	*/
}

// IsValid returns true if CR is semantically correct
func (r *ReconcileProjectIgnition) IsValid(obj metav1.Object) (bool, error) {
	cr, ok := obj.(*lbroudouxv1beta1.ProjectIgnition)
	// validation logic
	if !ok {
		return false, errors.New("Not a ProjectIgnition object")
	}
	if len(cr.Spec.Namespaces.Lifecycle) == len(cr.Spec.Namespaces.Definitions) {
		return true, nil
	}
	return false, errors.New("Not valid because lifecycles and defitions number does not match")
}

func (r *ReconcileProjectIgnition) manageCleanUpLogic(cr *lbroudouxv1beta1.ProjectIgnition) error {
	return nil
}

func (r *ReconcileProjectIgnition) manageOperatorLogic(cr *lbroudouxv1beta1.ProjectIgnition) error {

	for i := 0; i < len(cr.Spec.Namespaces.Definitions); i++ {
		var namespaceName = (cr.Spec.ProjectName + "-" + cr.Spec.Namespaces.Definitions[i].Name)

		// Check namespace already exist using a base v1 client.
		v1client, err := kubernetes.NewForConfig(r.restConfig)
		if err != nil {
			panic(err.Error())
		}
		namespace, err := v1client.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			log.Info("Namespace does not exists, creating it...")
			namespace = createNewNamespaceFromDefinition(&cr.Spec, &cr.Spec.Namespaces.Definitions[i])
			namespace, err = v1client.CoreV1().Namespaces().Create(namespace)
			if err != nil {
				log.Error(err, "Failed to create a Namespace")
				return err
			}

			// Set ownership of namespace to this CR
			if err := controllerutil.SetControllerReference(cr, namespace, r.scheme); err != nil {
				return err
			}

			if cr.Spec.Namespaces.Definitions[i].Roles != nil {
				log.Info("Now dealing with roles...")
				for j := 0; j < len(cr.Spec.Namespaces.Definitions[i].Roles); j++ {
					role := createRoleBindingFromDefinition(&cr.Spec, &cr.Spec.Namespaces.Definitions[i], &cr.Spec.Namespaces.Definitions[i].Roles[j])
					err = r.client.Create(context.TODO(), role)
					if err != nil {
						log.Error(err, "Failed to create a RoleBinding")
						return err
					}

					// Set ownership of RoleBinding to this CR
					if err := controllerutil.SetControllerReference(cr, role, r.scheme); err != nil {
						return err
					}
				}
			}

			if cr.Spec.Namespaces.Definitions[i].Quotas != nil {
				log.Info("Now dealing with quotas...")
			}
		} else {
			log.Info("Namespace " + namespace.Name + " already exists")
		}
	}

	// The discovery package is used to discover APIs supported by a Kubernetes API server.
	//r.GetDiscoveryClient()

	return nil
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

func createRoleBindingFromDefinition(spec *lbroudouxv1beta1.ProjectIgnitionSpec, def *lbroudouxv1beta1.DefinitionSpec, roleSpec *lbroudouxv1beta1.RoleSpec) *rbacv1.RoleBinding {
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

func replaceProjectOrNamespaceInString(projectName string, namespaceName string, stringToReplace string) string {
	if strings.Contains(stringToReplace, "{project}") {
		stringToReplace = strings.Replace(stringToReplace, "{project}", projectName, -1)
	}
	if strings.Contains(stringToReplace, "{namespace}") {
		stringToReplace = strings.Replace(stringToReplace, "{namespace}", namespaceName, -1)
	}
	return stringToReplace
}
