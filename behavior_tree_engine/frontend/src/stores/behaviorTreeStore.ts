import { defineStore } from 'pinia';
import type { VueFlowStore } from '@vue-flow/core'; // Using VueFlowStore type for elements

// Define interfaces for our domain models, matching backend structures eventually
// These might be simplified versions for now or live in a separate types/interfaces file later

export interface BehaviorTreeNodeData {
  label: string;
  // Add other node-specific data here, e.g., script, parameters
}

export interface BehaviorTreeNode extends VueFlowStore.Node<BehaviorTreeNodeData> {
  // Custom properties for your node type if needed
}

export interface BehaviorTreeEdge extends VueFlowStore.Edge {
  // Custom properties for your edge type if needed
}

export interface BehaviorTreeDefinition {
  nodes: BehaviorTreeNode[];
  edges: BehaviorTreeEdge[];
  // viewport can also be part of the definition if you want to save/restore it
  viewport?: VueFlowStore.Viewport;
}

export interface BehaviorTree {
  id: number;
  name: string;
  definition: BehaviorTreeDefinition; // Storing parsed definition
  rawDefinition: string; // Storing the raw JSON string from backend
  created_at: string;
  updated_at: string;
}


interface BehaviorTreeState {
  behaviorTrees: BehaviorTree[];
  selectedTree: BehaviorTree | null;
  isLoading: boolean;
  error: string | null;
}

export const useBehaviorTreeStore = defineStore('behaviorTree', {
  state: (): BehaviorTreeState => ({
    behaviorTrees: [],
    selectedTree: null,
    isLoading: false,
    error: null,
  }),
  getters: {
    getTreeById: (state) => (id: number) => {
      return state.behaviorTrees.find(tree => tree.id === id);
    },
    // Example: current tree elements for VueFlow
    currentVueFlowElements: (state): Array<BehaviorTreeNode | BehaviorTreeEdge> => {
      if (state.selectedTree && state.selectedTree.definition) {
        return [...state.selectedTree.definition.nodes, ...state.selectedTree.definition.edges];
      }
      return [];
    },
  },
  actions: {
    // Action to select a tree for editing
    selectTree(treeId: number | null) {
      if (treeId === null) {
        this.selectedTree = null;
        return;
      }
      const tree = this.behaviorTrees.find(t => t.id === treeId);
      if (tree) {
        // Attempt to parse the rawDefinition if it hasn't been parsed yet
        if (typeof tree.rawDefinition === 'string' && !tree.definition) {
          try {
            const parsedDefinition = JSON.parse(tree.rawDefinition) as BehaviorTreeDefinition;
            // Basic validation
            if (!parsedDefinition.nodes) parsedDefinition.nodes = [];
            if (!parsedDefinition.edges) parsedDefinition.edges = [];
            tree.definition = parsedDefinition;
          } catch (e) {
            console.error("Failed to parse tree definition:", e);
            this.error = "Failed to parse tree definition.";
            // Set a default empty definition to prevent errors
            tree.definition = { nodes: [], edges: [] };
          }
        } else if (!tree.definition) {
           // If rawDefinition is not a string or doesn't exist, and no definition exists
           tree.definition = { nodes: [], edges: [] };
        }
        this.selectedTree = tree;
      } else {
        this.error = `Tree with id ${treeId} not found.`;
        this.selectedTree = null;
      }
    },

    // Placeholder for fetching all trees from backend
    async fetchBehaviorTrees() {
      this.isLoading = true;
      this.error = null;
      try {
        // const response = await api.get('/behavior-trees'); // Replace with actual API call
        // this.behaviorTrees = response.data;
        // For now, using mock data:
        const mockTrees: BehaviorTree[] = [
          {
            id: 1, name: "Simple Test Tree", rawDefinition: JSON.stringify({ nodes: [{id: '1', type: 'custom', data: { label: 'Root Node'}, position: {x:100, y:100}, label: 'Root Node'}] , edges: [] }),
            definition: { nodes: [{id: '1', type: 'custom', data: { label: 'Root Node'}, position: {x:100, y:100}, label: 'Root Node'}] , edges: [] },
            created_at: new Date().toISOString(), updated_at: new Date().toISOString()
          },
          {
            id: 2, name: "Another Empty Tree", rawDefinition: JSON.stringify({ nodes: [], edges: [] }),
            definition: { nodes: [], edges: [] },
            created_at: new Date().toISOString(), updated_at: new Date().toISOString()
          }
        ];
        this.behaviorTrees = mockTrees.map(tree => {
          if (typeof tree.rawDefinition === 'string') {
            try {
              tree.definition = JSON.parse(tree.rawDefinition);
              if (!tree.definition.nodes) tree.definition.nodes = [];
              if (!tree.definition.edges) tree.definition.edges = [];
            } catch (e) {
              console.error("Error parsing mock tree definition", e);
              tree.definition = { nodes: [], edges: [] };
            }
          }
          return tree;
        });

      } catch (err) {
        this.error = 'Failed to fetch behavior trees.';
        console.error(err);
      } finally {
        this.isLoading = false;
      }
    },

    // Placeholder for creating a tree
    async createBehaviorTree(name: string, definition: BehaviorTreeDefinition) {
      this.isLoading = true;
      this.error = null;
      try {
        const rawDefinition = JSON.stringify(definition);
        // const response = await api.post('/behavior-trees', { name, definition: rawDefinition });
        // const newTree = response.data as BehaviorTree;
        // newTree.definition = definition; // Assign parsed def
        // this.behaviorTrees.push(newTree);
        // this.selectTree(newTree.id);

        // Mock creation:
        const newTree: BehaviorTree = {
          id: Date.now(), // Simple unique ID for mock
          name,
          rawDefinition,
          definition,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        this.behaviorTrees.push(newTree);
        this.selectTree(newTree.id);
        return newTree;

      } catch (err) {
        this.error = 'Failed to create behavior tree.';
        console.error(err);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    // Action to update nodes and edges for the selected tree
    updateSelectedTreeElements(elements: Array<BehaviorTreeNode | BehaviorTreeEdge>) {
      if (this.selectedTree && this.selectedTree.definition) {
        this.selectedTree.definition.nodes = elements.filter(el => 'position' in el) as BehaviorTreeNode[];
        this.selectedTree.definition.edges = elements.filter(el => 'source' in el) as BehaviorTreeEdge[];
        // Optionally, update rawDefinition immediately or on save
        this.selectedTree.rawDefinition = JSON.stringify(this.selectedTree.definition);
      }
    },

    updateNodeData(nodeId: string, data: Partial<BehaviorTreeNodeData>) {
        if (this.selectedTree && this.selectedTree.definition) {
            const node = this.selectedTree.definition.nodes.find(n => n.id === nodeId);
            if (node) {
                node.data = { ...node.data, ...data };
                // Update rawDefinition as well
                this.selectedTree.rawDefinition = JSON.stringify(this.selectedTree.definition);
            }
        }
    }
  },
});
