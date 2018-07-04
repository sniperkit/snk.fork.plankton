package task

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"sync"
)

type TaskRunner interface {
	Run()
	// requires TaskRuner to have an embedded Task
	GetTask() *Task
}

type Task struct {
	// TODO: does children need to be TaskRunner or can I get away with making everything
	//       a task
	Name           string
	Children       []TaskRunner
	Parent         TaskRunner
	ResultsChannel chan string
	State          string
	Params         []*TaskParam
}

type TaskParam struct {
	Name string
	Data reflect.Value
}

func NewTask(name string, children []TaskRunner, parent TaskRunner) *Task {
	return &Task{
		Name:           name,
		Children:       children,
		Parent:         parent,
		ResultsChannel: make(chan string),
		State:          "waiting",
		Params:         []*TaskParam{},
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()

}

// TODO: Consider getting rid of this
//func (ts *Task) AddChild(child TaskRunner) []TaskRunner {
//	ts.Children = append(ts.Children, child)
//	return ts.Children
//}

func (ts *Task) GetHash() string {
	param_strings := []string{}
	for _, parm := range ts.Params {
		param_strings = append(param_strings, fmt.Sprintf("%s:%s", parm.Name, parm.Data))
	}
	param_string := strings.Join(param_strings, "_")
	param_hash := hash(param_string)
	hash_elements := []string{
		ts.Name,
		param_string,
		string(param_hash),
	}
	return strings.Join(hash_elements, "_")
}

// Uses reflection to inspect struct elements for 'task_param' tag
// and sets tr.Task.Params accordingly
func SetTaskParams(tr TaskRunner) ([]*TaskParam, error) {

	const TASK_PARAM_TAG = "task_param"
	var task_params []*TaskParam

	v := reflect.ValueOf(tr).Elem()
	for i := 0; i < v.NumField(); i++ {
		field_info := v.Type().Field(i)
		tag := field_info.Tag
		_, ok := tag.Lookup(TASK_PARAM_TAG)
		if ok {
			new_param := TaskParam{
				Name: field_info.Name,
				Data: getFieldValue(tr, field_info.Name),
			}
			task_params = append(task_params, &new_param)
		}
	}

	tr.GetTask().Params = task_params
	return tr.GetTask().Params, nil
}

func getFieldValue(tr TaskRunner, field_name string) reflect.Value {
	// TODO: check this works on different types other than int
	tr_reflect := reflect.ValueOf(tr)
	field_val := reflect.Indirect(tr_reflect).FieldByName(field_name)
	return field_val

}

// Given Params and an empty Foo, Returns a new Foo.
// Note that a Foo is an interface, for this program to
// work the struct that satifies the Foo interface must be passed to CreateFooFromParams
// as a pointer.  See this article for a thorough explanation:
// https://stackoverflow.com/questions/6395076/using-reflect-how-do-you-set-the-value-of-a-struct-field
// http://speakmy.name/2014/09/14/modifying-interfaced-go-struct/
func CreateTaskRunnerFromParams(tr TaskRunner, params []*TaskParam) error {
	stype := reflect.ValueOf(tr).Elem()

	param_name_value_map := map[string]reflect.Value{}
	for _, param := range params {
		param_name_value_map[param.Name] = param.Data
	}

	if stype.Kind() == reflect.Struct {
		for name, val := range param_name_value_map {
			f := stype.FieldByName(name)
			if f.CanSet() {
				// TODO: support more kinds of fields
				switch f.Kind() {
				case reflect.Int:
					f.SetInt(val.Int())
				case reflect.String:
					f.SetString(val.String())
				default:
					// TODO: think about what to do in this case
					return fmt.Errorf("%s not supported as TaskParam yet!", f.Kind())
				}

			} else {
				fmt.Printf("Cannot set %s %v\n", name, f)
			}
		}
	}

	return nil

}

func (ts *Task) SetState(new_state string) (string, error) {
	valid_states := []string{
		"waiting",
		"running",
		"complete",
	}

	valid_state_param := false
	for _, state := range valid_states {
		if new_state == state {
			valid_state_param = true
			break
		}

	}

	if !valid_state_param {
		return "", fmt.Errorf("Invalid state on task {}", new_state)
	}

	ts.State = new_state
	return ts.State, nil
}

// Runs a TaskRunner, sets state and notifies waiting group when run is done
func RunTaskRunner(tsk_runner TaskRunner, wg *sync.WaitGroup) {
	defer wg.Done()
	tsk_runner.GetTask().SetState("running")
	fmt.Printf("Running Task: %s\n", tsk_runner.GetTask().Name)
	tsk_runner.Run()
	tsk_runner.GetTask().SetState("complete")
}

// TODO: do a better job detecting the D part

// TODO: we also need some way of creating a map of tasks
//       to check acyclical property
//       Going to implement task_has on Task
func VerifyDAG(root_task *Task) bool {

	task_set := make(map[string]struct{})

	task_queue := []*Task{}
	task_queue = append(task_queue, root_task)
	for len(task_queue) > 0 {
		curr := task_queue[0]
		task_queue = task_queue[1:]

		_, ok := task_set[curr.GetHash()]
		if ok {
			return false
		} else {
			task_set[curr.GetHash()] = struct{}{}
		}
		for _, child := range curr.Children {
			task_queue = append(task_queue, child.GetTask())
		}

	}

	return true
}
