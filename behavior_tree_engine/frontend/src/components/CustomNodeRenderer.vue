<template>
  <div :class="['custom-node', `node-type-${props.node.data.engineType}` , { selected: props.node.selected }]" @dblclick="handleDoubleClick">
    <div class="node-header">
      <span class="node-icon mr-2">{{ getNodeIcon(props.node.data.engineType) }}</span>
      <strong class="node-label">{{ props.node.data.label || props.node.label || 'Unnamed Node' }}</strong>
    </div>
    <div v-if="props.node.data.script !== undefined" class="node-body">
      <p class="text-xs text-gray-400 truncate" title="Script content (preview)">
        Script: {{ props.node.data.script || '(empty)' }}
      </p>
    </div>
    <div v-if="props.node.data.engineType === 'sequence' || props.node.data.engineType === 'selector'" class="node-body">
        <p class="text-xs text-gray-400 italic">Composite Node</p>
    </div>
     <div v-if="props.node.data.engineType === 'inverter' || props.node.data.engineType === 'succeeder' || props.node.data.engineType === 'failer'" class="node-body">
        <p class="text-xs text-gray-400 italic">Decorator Node</p>
    </div>

    <Handle type="target" :position="Position.Left" :style="{ background: '#555' }" />
    <Handle type="source" :position="Position.Right" :style="{ background: '#555' }" />

    <button
      v-if="props.node.data.script !== undefined"
      @click.stop="openScriptEditor"
      class="edit-button text-xs bg-cyan-600 hover:bg-cyan-500 text-white py-1 px-2 rounded mt-2 w-full">
      Edit Script
    </button>

    <button
      @click.stop="editLabel"
      class="edit-label-button text-xs bg-blue-600 hover:bg-blue-500 text-white py-1 px-2 rounded mt-1 w-full">
      Edit Label
    </button>

    <ScriptEditorModal
      :visible="isModalVisible"
      :initial-script="currentScriptContent"
      :title="`Edit Script for ${props.node.data.label || 'Node'}`"
      @update:visible="isModalVisible = $event"
      @save="handleScriptSave"
    />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position, NodeProps } from '@vue-flow/core';
import { ref } from 'vue';
import { useBehaviorTreeStore, BehaviorTreeNodeData } from '../stores/behaviorTreeStore';
import ScriptEditorModal from './ScriptEditorModal.vue';

const props = defineProps<{
  // node: NodeProps<BehaviorTreeNodeData> - VueFlow provides `id`, `type`, `label`, `selected`, `data`, `events`, etc.
  // We only need to define the structure of `data` if it's more specific than Record<string, any>
  // However, to get full type inference on `props.node.data` it's better to use NodeProps
   node: NodeProps<BehaviorTreeNodeData & { engineType?: string; script?: string }>;
}>();

const store = useBehaviorTreeStore();
const isModalVisible = ref(false);
const currentScriptContent = ref('');

const getNodeIcon = (engineType: string | undefined) => {
  switch (engineType) {
    case 'action': return '⚡';
    case 'condition': return '❓';
    case 'sequence': return '➡️';
    case 'selector': return '⚖️';
    case 'inverter': return '🔄';
    case 'succeeder': return '✅';
    case 'failer': return '❌';
    default: return '📦';
  }
};

const openScriptEditor = () => {
  currentScriptContent.value = props.node.data.script || '';
  isModalVisible.value = true;
};

const handleDoubleClick = () => {
  if (props.node.data.script !== undefined) {
    openScriptEditor();
  } else {
    editLabel();
  }
};

const handleScriptSave = (newScript: string) => {
  store.updateNodeData(props.node.id, { script: newScript });
  isModalVisible.value = false; // Modal closes itself, but good to ensure state consistency
};

const editLabel = () => {
  const newLabel = prompt(`Edit label for node "${props.node.data.label || props.node.id}":`, props.node.data.label || '');
  if (newLabel !== null && newLabel !== props.node.data.label) {
    store.updateNodeData(props.node.id, { label: newLabel });
    // Note: VueFlow's label prop on the node itself might need updating if you rely on it directly for rendering.
    // If `elements` in BehaviorTreeEditor.vue is correctly watching the store and reconstructing nodes, this should be fine.
    // Alternatively, emit an event to parent to update the node's top-level label if necessary.
  }
};

</script>

<style scoped>
.custom-node {
  background-color: #2d3748; /* bg-gray-700 */
  border: 1px solid #4a5568; /* border-gray-600 */
  color: #e2e8f0; /* text-gray-300 */
  border-radius: 6px;
  padding: 8px 12px;
  min-width: 150px;
  font-family: Arial, sans-serif;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.custom-node.selected {
  border-color: #38bdf8; /* border-cyan-500 */
  box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.5), 0 4px 8px rgba(0,0,0,0.3);
}

.node-header {
  display: flex;
  align-items: center;
  font-size: 1rem;
  margin-bottom: 4px;
  color: #90cdf4; /* blue-300 for header text */
}

.node-label {
  font-weight: bold;
}

.node-body {
  font-size: 0.8rem;
  color: #a0aec0; /* gray-500 */
  margin-top: 4px;
}

.edit-button {
  display: block; /* Make button take its own line or inline-block for side-by-side */
  width: calc(100% - 0px); /* Full width button */
  box-sizing: border-box;
}

/* Specific styles based on engineType if needed */
.node-type-action { border-left: 3px solid #f59e0b; /* amber-500 */ }
.node-type-condition { border-left: 3px solid #ec4899; /* pink-500 */ }
.node-type-sequence { border-left: 3px solid #84cc16; /* lime-500 */ }
.node-type-selector { border-left: 3px solid #6366f1; /* indigo-500 */ }
.node-type-inverter { border-left: 3px solid #a855f7; /* purple-500 */}
.node-type-succeeder { border-left: 3px solid #22c55e; /* green-500 */}
.node-type-failer { border-left: 3px solid #ef4444; /* red-500 */}

</style>
