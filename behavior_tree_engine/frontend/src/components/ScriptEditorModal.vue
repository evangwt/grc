<template>
  <div v-if="visible" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content bg-gray-800 text-white rounded-lg shadow-xl p-6 w-full max-w-2xl">
      <h3 class="text-xl font-semibold mb-4 text-cyan-400">{{ title }}</h3>

      <div ref="editorContainer" class="editor-container w-full h-80 border border-gray-700 rounded">
        <!-- Monaco Editor will be mounted here -->
      </div>

      <div class="modal-actions mt-6 flex justify-end space-x-3">
        <button @click="closeModal" class="px-4 py-2 bg-gray-600 hover:bg-gray-500 rounded-md text-white transition-colors">
          Cancel
        </button>
        <button @click="saveScript" class="px-4 py-2 bg-green-500 hover:bg-green-600 rounded-md text-white font-bold transition-colors">
          Save Script
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import * as monaco from 'monaco-editor';

// Define props and emits
const props = defineProps<{
  visible: boolean;
  initialScript: string;
  title?: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'save', script: string): void;
}>();

const editorContainer = ref<HTMLElement | null>(null);
let editorInstance: monaco.editor.IStandaloneCodeEditor | null = null;

const closeModal = () => {
  emit('update:visible', false);
};

const saveScript = () => {
  if (editorInstance) {
    emit('save', editorInstance.getValue());
  }
  closeModal();
};

watch(() => props.visible, (newValue) => {
  if (newValue) {
    nextTick(() => { // Ensure DOM is ready
      if (editorContainer.value && !editorInstance) {
        monaco.editor.defineTheme('custom-dark', {
            base: 'vs-dark', // can be vs, vs-dark or hc-black
            inherit: true, // can also be false to completely replace the base theme
            rules: [
                // { token: 'comment', foreground: 'ffa500', fontStyle: 'italic underline' },
                // { token: 'comment.js', foreground: '008800', fontStyle: 'bold' },
                // { token: 'comment.css', foreground: '0000ff' } // will inherit fontStyle from `comment`
            ],
            colors: {
                'editor.background': '#1f2937', // bg-gray-800 (a bit darker for editor)
                'editor.foreground': '#d1d5db', // text-gray-300
                'editorLineNumber.foreground': '#4b5563', // gray-600
                'editorLineNumber.activeForeground': '#9ca3af', // gray-400
                'editorCursor.foreground': '#67e8f9', // cyan-300
                // ... more theme color overrides
            }
        });
        monaco.editor.setTheme('custom-dark');

        editorInstance = monaco.editor.create(editorContainer.value, {
          value: props.initialScript,
          language: 'javascript',
          theme: 'custom-dark', // vs-dark or your custom theme
          automaticLayout: true, // Adjusts editor layout on container resize
          fontSize: 14,
          minimap: { enabled: true },
        });
      } else if (editorInstance) {
        // If editor already exists and modal becomes visible again, update its content
        editorInstance.setValue(props.initialScript);
      }
       // Focus the editor when modal becomes visible
      setTimeout(() => editorInstance?.focus(), 100);
    });
  } else {
    // Optional: Destroy editor when not visible to save resources,
    // but this means it needs to be recreated each time.
    // For simplicity, we are not destroying it here, just hiding.
    // If performance becomes an issue, consider disposing:
    // if (editorInstance) {
    //   editorInstance.dispose();
    //   editorInstance = null;
    // }
  }
});

onMounted(() => {
  // Monaco workers setup (important for some features like web workers for intellisense)
  // This path is relative to the `public` directory or where Vite serves static assets from.
  // You might need to copy Monaco's worker files to your public directory.
  // This is a common setup but might need adjustment based on your Vite config.
  // (self as any).MonacoEnvironment = {
  //   getWorkerUrl: function (_moduleId: any, label: string) {
  //     if (label === 'json') {
  //       return './json.worker.bundle.js';
  //     }
  //     if (label === 'css' || label === 'scss' || label === 'less') {
  //       return './css.worker.bundle.js';
  //     }
  //     if (label === 'html' || label === 'handlebars' || label === 'razor') {
  //       return './html.worker.bundle.js';
  //     }
  //     if (label === 'typescript' || label === 'javascript') {
  //       return './ts.worker.bundle.js';
  //     }
  //     return './editor.worker.bundle.js';
  //   }
  // };
});


onBeforeUnmount(() => {
  if (editorInstance) {
    editorInstance.dispose();
    editorInstance = null;
  }
});

</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000; /* Ensure it's on top */
}

.modal-content {
  min-height: 300px; /* Ensure modal has a decent size */
  display: flex;
  flex-direction: column;
}

.editor-container {
  flex-grow: 1; /* Make editor take available space */
  min-height: 200px; /* Minimum height for editor area */
}
</style>
