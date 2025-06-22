<template>
  <aside class="node-palette bg-gray-750 p-3 rounded-lg shadow-lg w-64 h-full overflow-y-auto">
    <h3 class="text-lg font-semibold text-cyan-300 mb-3 border-b border-gray-600 pb-2">Node Types</h3>
    <div
      v-for="nodeType in availableNodeTypes"
      :key="nodeType.type"
      class="palette-item bg-gray-700 p-3 mb-2 rounded-md shadow hover:bg-gray-600 cursor-grab transition-colors duration-150"
      draggable="true"
      @dragstart="onDragStart($event, nodeType)"
    >
      <p class="font-medium text-gray-100">{{ nodeType.label }}</p>
      <p class="text-xs text-gray-400">{{ nodeType.description }}</p>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue';

export interface PaletteNodeType {
  type: string; // Unique identifier for the node type, e.g., 'sequenceNode', 'selectorNode', 'actionNode'
  label: string; // User-friendly label, e.g., "Sequence"
  description: string; // Short description
  defaultData?: Record<string, any>; // Default data for the node when created
  vueFlowNodeType?: string; // Optional: if you use custom node types in VueFlow, e.g., 'customInput'
}

const availableNodeTypes = ref<PaletteNodeType[]>([
  {
    type: 'sequence',
    label: 'Sequence',
    description: 'Executes children in order. Succeeds if all succeed.',
    vueFlowNodeType: 'default', // or your custom node type for composites
    defaultData: { label: 'Sequence' }
  },
  {
    type: 'selector',
    label: 'Selector',
    description: 'Executes children until one succeeds.',
    vueFlowNodeType: 'default',
    defaultData: { label: 'Selector' }
  },
  {
    type: 'action',
    label: 'Action',
    description: 'Performs a task. (e.g. run script)',
    vueFlowNodeType: 'custom', // Example: if actions have a special rendering/handling
    defaultData: { label: 'Action', script: 'console.log("Hello from Action!");' }
  },
  {
    type: 'condition',
    label: 'Condition',
    description: 'Checks a condition. (e.g. run script)',
    vueFlowNodeType: 'custom', // Example: if conditions have special rendering
    defaultData: { label: 'Condition', script: 'return true;' }
  },
  {
    type: 'inverter',
    label: 'Inverter',
    description: 'Inverts the result of its child. (Decorator)',
    vueFlowNodeType: 'default',
    defaultData: { label: 'Inverter' }
  },
  {
    type: 'succeeder',
    label: 'Succeeder',
    description: 'Always returns SUCCESS, regardless of child. (Decorator)',
    vueFlowNodeType: 'default',
    defaultData: { label: 'Succeeder' }
  },
  {
    type: 'failer',
    label: 'Failer',
    description: 'Always returns FAILURE, regardless of child. (Decorator)',
    vueFlowNodeType: 'default',
    defaultData: { label: 'Failer' }
  }
  // Add more node types here (e.g., Inverter, Succeeder, Failer, custom script nodes)
]);

const onDragStart = (event: DragEvent, nodeType: PaletteNodeType) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/json', JSON.stringify(nodeType));
    event.dataTransfer.effectAllowed = 'copy';
    console.log(`Dragging node type: ${nodeType.label}`);
  }
};
</script>

<style scoped>
.node-palette {
  /* Custom scrollbar for webkit browsers */
  &::-webkit-scrollbar {
    width: 6px;
  }
  &::-webkit-scrollbar-track {
    background: #374151; /* bg-gray-700 */
  }
  &::-webkit-scrollbar-thumb {
    background: #4b5563; /* bg-gray-600 */
    border-radius: 3px;
  }
  &::-webkit-scrollbar-thumb:hover {
    background: #6b7280; /* bg-gray-500 */
  }
}

.palette-item {
  border-left: 4px solid #38bdf8; /* cyan-500 */
}
.palette-item:hover {
  border-left-color: #0ea5e9; /* cyan-600 */
}
</style>
