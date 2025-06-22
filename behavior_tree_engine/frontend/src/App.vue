<script setup lang="ts">
import EditorView from './views/EditorView.vue';
</script>

<template>
  <div id="app-container" class="dark bg-gray-900 text-white h-screen flex flex-col">
    <header class="p-3 bg-gray-800 shadow-md flex-shrink-0">
      <h1 class="text-3xl font-bold text-cyan-400 text-center">Behavior Tree Engine</h1>
    </header>

    <main class="flex-grow overflow-auto">
      <EditorView />
    </main>

    <footer class="p-2 bg-gray-800 text-center text-xs text-gray-500 flex-shrink-0">
      Powered by Vue, Golang, VueFlow, Monaco & TailwindCSS
    </footer>
  </div>
</template>

<style>
/* Ensure html and body take full height for the dark theme */
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}

#app-container {
 /* Ensures the container itself uses the full viewport height if #app is the direct parent */
  height: 100%;
}

/* Global scrollbar styling for a more integrated dark theme (optional) */
/* For Webkit browsers (Chrome, Safari, newer Edge) */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1f2937; /* bg-gray-800 */
}

::-webkit-scrollbar-thumb {
  background: #4b5563; /* bg-gray-600 */
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #6b7280; /* bg-gray-500 */
}

/* For Firefox */
/* Firefox scrollbar styling is less direct, often needs browser settings or more complex CSS */
/* For basic color, can try: */
/*
* {
  scrollbar-width: thin;
  scrollbar-color: #4b5563 #1f2937;
}
*/
</style>
