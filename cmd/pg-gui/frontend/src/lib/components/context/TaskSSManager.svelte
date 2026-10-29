<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<!-- @component
register the stream pipes of TaskService,and register callback on data pkg
 -->
<script lang="ts" module>
	import type { TraceView } from '@pg/protocol';
	import { TaskSStream, type Task, type TaskResMeta } from '@services';
	import { Stream } from '@wailsio/runtime';

	type Handle<T> = (v: T) => void;
	const taskStream = Stream(TaskSStream.TaskSSTask);
	const taskHandlers = new Set<Handle<Task>>();
	taskStream.onmessage = (ev) => {
		if (ev.data) {
			taskHandlers.forEach((call) => {
				call(ev.data as Task);
			});
		}
	};
	const resMetaStream = Stream(TaskSStream.TaskSSResMeta);
	const resMetaHandlers = new Set<Handle<TaskResMeta>>();
	resMetaStream.onmessage = (ev) => {
		if (ev.data) {
			resMetaHandlers.forEach((fn) => {
				fn(ev.data as TaskResMeta);
			});
		}
	};
	const traceViewStream = Stream(TaskSStream.TaskSSTraceView);
	const traceViewHandlers = new Set<Handle<TraceView>>();
	traceViewStream.onmessage = (ev) => {
		if (ev.data) {
			traceViewHandlers.forEach((fn) => {
				fn(ev.data as TraceView);
			});
		}
	};

	export const TaskSSReg = {
		onTask: (call: Handle<Task>) => {
			taskHandlers.add(call);
			return () => {
				taskHandlers.delete(call);
			};
		},
		onTaskResMeta: (call: Handle<TaskResMeta>) => {
			resMetaHandlers.add(call);
			return () => {
				resMetaHandlers.delete(call);
			};
		},
		onTrace: (call: Handle<TraceView>) => {
			traceViewHandlers.add(call);
			return () => {
				traceViewHandlers.delete(call);
			};
		},
	};
</script>
