<template>
  <div class="editor-view-container flex flex-col h-full p-1 sm:p-4 gap-2 sm:gap-4">
    <div class="controls-bar bg-gray-800 p-3 rounded-lg shadow-md flex flex-col sm:flex-row items-center justify-between gap-2">
      <h2 class="text-xl sm:text-2xl font-semibold text-cyan-400">Behavior Tree Editor</h2>
      <div class="flex items-center gap-2">
        <select
          v-if="store.behaviorTrees.length > 0"
          @change="handleTreeSelection"
          class="bg-gray-700 text-white p-2 rounded-md border border-gray-600 focus:ring-cyan-500 focus:border-cyan-500 text-sm sm:text-base">
          <option :value="undefined" disabled :selected="!store.selectedTree">Select a Tree</option>
          <option v-for="tree in store.behaviorTrees" :key="tree.id" :value="tree.id" :selected="store.selectedTree?.id === tree.id">
            {{ tree.name }} (ID: {{ tree.id }})
          </option>
        </select>
        <button @click="createNewTree" class="bg-green-500 hover:bg-green-600 text-white font-bold py-2 px-3 sm:px-4 rounded text-sm sm:text-base">
          New Tree
        </button>
        <!-- More controls can be added here: save, load, etc. -->
      </div>
    </div>

    <div class="editor-layout flex flex-grow gap-2 sm:gap-4 overflow-hidden">
      <NodePalette class="flex-shrink-0 w-48 sm:w-64" />
      <div class="editor-area flex-grow h-full min-w-0"> <!-- min-w-0 is important for flex item to shrink properly -->
        <BehaviorTreeEditor v-if="store.selectedTree" class="h-full w-full" />
        <div v-else class="flex items-center justify-center h-full bg-gray-800 rounded-lg">
          <p class="text-gray-500 text-lg sm:text-xl p-4 text-center">Select or create a behavior tree to start editing.</p>
        </div>
      </div>
    </div>

    <!-- Simple selected tree info for debugging -->
    <div v-if="store.selectedTree" class="p-2 bg-gray-700 rounded text-xs mt-1 sm:mt-2 text-gray-300">
      <p>Selected: <span class="font-semibold">{{ store.selectedTree.name }}</span></p>
      <p>Nodes: <span class="font-semibold">{{ store.selectedTree.definition.nodes.length }}</span>, Edges: <span class="font-semibold">{{ store.selectedTree.definition.edges.length }}</span></p>
    </div>

  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue';
import BehaviorTreeEditor from '../components/BehaviorTreeEditor.vue';
import NodePalette from '../components/NodePalette.vue';
import { useBehaviorTreeStore } from '../stores/behaviorTreeStore';

const store = useBehaviorTreeStore();

// Use a computed property for the select binding to avoid issues with initial undefined value
const selectedTreeId = computed({
  get: () => store.selectedTree?.id,
  set: (value) => {
    if (value !== undefined) {
      store.selectTree(value);
    } else {
      store.selectTree(null);
    }
  }
});


onMounted(async () => {
  if (store.behaviorTrees.length === 0) {
    await store.fetchBehaviorTrees();
  }
  // Auto-select first tree if none is selected and trees are available
  // This logic is also present in BehaviorTreeEditor, but can be here too for initial view setup
  if (!store.selectedTree && store.behaviorTrees.length > 0) {
     store.selectTree(store.behaviorTrees[0].id);
  }
});

const handleTreeSelection = (event: Event) => {
  const target = event.target as HTMLSelectElement;
  // Value might be an empty string if a "Select a tree" option is re-selected somehow, parse carefully
  const treeId = target.value && target.value !== "undefined" ? parseInt(target.value, 10) : null;
  store.selectTree(treeId);
};

const createNewTree = async () => {
  const treeName = prompt("Enter the name for the new behavior tree:", "New Tree " + (store.behaviorTrees.length + 1));
  if (treeName) {
    const newTree = await store.createBehaviorTree(treeName, { nodes: [], edges: [] });
    if (newTree) {
        // The store action should ideally handle selecting the new tree.
        // If not, explicitly select it here:
        // store.selectTree(newTree.id);
    }
  }
};

</script>

<style scoped>
.editor-view-container {
  background-color: #1a202c; /* Slightly darker than gray-800 for main bg */
  color: #e5e7eb; /* text-gray-200 */
}

.editor-layout {
  /* This will contain the palette and the editor area */
}

.editor-area {
  /* Ensure this area can shrink and grow correctly and has a minimum size */
  min-height: 300px; /* Example minimum height */
}

/* Responsive adjustments if needed */
@media (max-width: 640px) { /* sm breakpoint */
  .controls-bar {
    /* Stack controls vertically on small screens */
  }
  .editor-layout {
    flex-direction: column; /* Stack palette on top of editor on very small screens if desired */
  }
  /* Adjust palette width if it stacks or becomes too dominant */
}
</style>
