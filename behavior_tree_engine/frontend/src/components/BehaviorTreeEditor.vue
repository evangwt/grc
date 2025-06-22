<template>
  <div
    class="vue-flow-wrapper w-full h-full bg-gray-800 border border-gray-700 rounded-lg"
    @dragover.prevent="onDragOver"
    @drop.prevent="onDrop"
  >
    <VueFlow
      v-model="elements"
      :fit-view-on-init="true"
      :nodes-draggable="true"
      :edges-updatable="true"
      :connectable="true"
      @nodes-change="onNodesChangeHandler"
      @edges-change="onEdgesChangeHandler"
      @connect="onConnectHandler"
    >
      <Background :variant="BackgroundVariant.Dots" :gap="20" :size="1" color="#555" />
      <Controls />
      <MiniMap />

      <!-- Custom Node Slot for specific node types -->
      <template #node-custom="props">
        <CustomNodeRenderer :node="props" />
      </template>
    </VueFlow>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';
import { VueFlow, useVueFlow, Connection, Edge, Node, Elements, FlowEvents, XYPosition, Viewport } from '@vue-flow/core';
import { Background, BackgroundVariant } from '@vue-flow/background';
import { Controls } from '@vue-flow/controls';
import { MiniMap } from '@vue-flow/minimap';
import { useBehaviorTreeStore, BehaviorTreeNode, BehaviorTreeEdge, BehaviorTreeNodeData } from '../stores/behaviorTreeStore';
import type { PaletteNodeType } from './NodePalette.vue'; // Assuming NodePalette exports this
import CustomNodeRenderer from './CustomNodeRenderer.vue';


import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css'; // Default theme for nodes
import '@vue-flow/controls/dist/style.css';
import '@vue-flow/minimap/dist/style.css';
import '@vue-flow/background/dist/style.css';

const store = useBehaviorTreeStore();
const {
  addEdges,
  addNodes: vueFlowAddNodes, // Renamed to avoid conflict if store had addNodes
  onNodesChange,
  onEdgesChange,
  onConnect,
  project, // For converting screen coords to flow coords
  vueFlowInstance // Access to the VueFlow instance
} = useVueFlow();

const elements = ref<Elements>([]);
let internalUpdate = false; // Flag to prevent feedback loops in watchers

watch(() => store.selectedTree, (currentTree) => {
  if (internalUpdate) return;

  if (currentTree && currentTree.definition) {
    console.log("Editor: Selected tree changed. New definition:", JSON.parse(JSON.stringify(currentTree.definition)));
    const nodesWithCorrectType = currentTree.definition.nodes.map(node => ({
      ...node,
      label: node.data?.label || node.id,
      type: node.type || 'default', // Ensure type is set, default for VueFlow
    }));
    internalUpdate = true;
    elements.value = [...nodesWithCorrectType, ...currentTree.definition.edges];
    nextTick(() => internalUpdate = false);
  } else {
    console.log("Editor: No tree selected or definition missing.");
    internalUpdate = true;
    elements.value = [];
    nextTick(() => internalUpdate = false);
  }
}, { immediate: true, deep: true });


const onNodesChangeHandler = (flowEvents: FlowEvents.NodesChange) => {
  if (internalUpdate) return;
  onNodesChange(flowEvents); // Let VueFlow update its internal state and v-model (elements)
};

const onEdgesChangeHandler = (flowEvents: FlowEvents.EdgesChange) => {
  if (internalUpdate) return;
  onEdgesChange(flowEvents);
};

const onConnectHandler = (connection: Connection | Edge) => {
  if (internalUpdate) return;
  addEdges([connection as Edge]); // This will trigger the watch on elements
};

// Watch for changes in local elements (driven by VueFlow interactions or drops) and update the store
watch(elements, (newElements) => {
  if (internalUpdate) return;
  if (store.selectedTree) {
    console.log("Editor: Local elements changed, updating store.", JSON.parse(JSON.stringify(newElements)));
    internalUpdate = true; // Prevent this watch from re-triggering due to store update
    store.updateSelectedTreeElements(newElements as Array<BehaviorTreeNode | BehaviorTreeEdge>);
    nextTick(() => internalUpdate = false);
  }
}, { deep: true });


const onDragOver = (event: DragEvent) => {
  event.preventDefault(); // Necessary to allow dropping
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy';
  }
};

let idCounter = 0; // Simple ID counter for new nodes, ensure it's unique enough or use UUIDs
watch(() => store.selectedTree?.definition.nodes, (nodes) => {
    if (nodes && nodes.length > 0) {
        const maxId = nodes.reduce((max, node) => {
            const numericId = parseInt(node.id.split('_').pop() || '0', 10);
            return Math.max(max, numericId);
        }, 0);
        idCounter = maxId + 1;
    } else {
        idCounter = 1;
    }
}, {deep: true, immediate: true});


const onDrop = (event: DragEvent) => {
  if (!event.dataTransfer || !store.selectedTree || !vueFlowInstance.value) return;

  const typeDataString = event.dataTransfer.getData('application/json');
  if (!typeDataString) return;

  const paletteNode = JSON.parse(typeDataString) as PaletteNodeType;

  // Get drop position relative to the pane
  const { x, y } = vueFlowInstance.value.screenToFlowCoordinate({
    x: event.clientX,
    y: event.clientY,
  });

  const newNodeId = `${paletteNode.type}_${idCounter++}`;

  const newNode: BehaviorTreeNode = {
    id: newNodeId,
    type: paletteNode.vueFlowNodeType || 'default', // Use 'custom' for CustomNodeRenderer or specific types
    label: paletteNode.label, // Initial label
    position: { x, y },
    data: {
        label: paletteNode.label, // Data for custom node
        ...(paletteNode.defaultData || {}),
        // Ensure any specific fields for different node types are initialized here
        // For example, 'script' for action/condition nodes:
        script: paletteNode.defaultData?.script || (paletteNode.type === 'action' || paletteNode.type === 'condition' ? '' : undefined),
        // Store the conceptual type from palette for engine use
        engineType: paletteNode.type
    },
    // You might need to define source/target handle positions based on node type
    // sourcePosition: Position.Right,
    // targetPosition: Position.Left,
  };

  console.log("Dropping new node:", JSON.parse(JSON.stringify(newNode)));
  vueFlowAddNodes([newNode]); // This adds to local 'elements' and triggers the watch to update the store
};


onMounted(async () => {
  if (store.behaviorTrees.length === 0) {
    await store.fetchBehaviorTrees();
  }
  if (!store.selectedTree && store.behaviorTrees.length > 0) {
    store.selectTree(store.behaviorTrees[0].id);
  }
});

</script>

<style scoped>
.vue-flow-wrapper {
  min-height: 500px;
}

/* Theme overrides for VueFlow default nodes */
:deep(.vue-flow__node-default) {
  background-color: #374151; /* bg-gray-700 */
  border: 1px solid #4b5563; /* border-gray-600 */
  color: #e5e7eb; /* text-gray-200 */
  border-radius: 0.375rem; /* rounded-md */
  padding: 0.5rem 0.75rem;
}
:deep(.vue-flow__node-default.selected) {
  border-color: #0ea5e9; /* border-cyan-500 */
  box-shadow: 0 0 0 2px #0ea5e9;
}

/* Handle styles */
:deep(.vue-flow__handle) {
  background-color: #60a5fa; /* bg-blue-400 */
  border: 1px solid #3b82f6; /* border-blue-500 */
  width: 10px;
  height: 10px;
}
:deep(.vue-flow__handle-connecting) {
  background-color: #a5f3fc; /* cyan-200 */
}
:deep(.vue-flow__handle-valid) {
   background-color: #4ade80; /* green-400 */
}


/* Edge styles */
:deep(.vue-flow__edge-path) {
  stroke: #60a5fa; /* stroke-blue-400 */
  stroke-width: 2;
}
:deep(.vue-flow__edge.selected .vue-flow__edge-path) {
  stroke: #0ea5e9; /* cyan-500 */
}

/* Controls and Minimap styling from previous step can be kept or refined */
:deep(.vue-flow__controls) {
    background-color: rgba(31, 41, 55, 0.85); /* bg-gray-800 with opacity */
    border-radius: 0.5rem; /* rounded-lg */
    padding: 0.25rem;
}
:deep(.vue-flow__controls-button svg) {
    fill: #9ca3af; /* text-gray-400 */
}
:deep(.vue-flow__controls-button:hover svg) {
    fill: #e5e7eb; /* text-gray-200 */
}

:deep(.vue-flow__minimap) {
    background-color: rgba(31, 41, 55, 0.85); /* bg-gray-800 with opacity */
    border-radius: 0.5rem; /* rounded-lg */
}
:deep(.vue-flow__minimap-mask) {
    fill: rgba(75, 85, 99, 0.3); /* bg-gray-500 with opacity */
}
:deep(.vue-flow__minimap-node) {
    fill: #60a5fa; /* fill-blue-400 */
    stroke: none;
}

:deep(.vue-flow__background) {
    background-color: #1f2937; /* bg-gray-800 for the pattern background */
}
</style>
