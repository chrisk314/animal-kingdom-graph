<template>
  <div>
    <GraphVisualisation :data="graphData" />
  </div>
</template>

<script>
import GraphVisualisation from './components/GraphVisualisation.vue'

export default {
  name: 'App',
  components: {
    GraphVisualisation
  },
  data() {
    return {
      graphData: {
        elements: [],
        style: [
          {
            selector: 'node',
            style: {
              'background-color': '#666',
              label: 'data(name)',
            },
          },
          {
            selector: 'edge',
            style: {
              'line-color': '#ccc',
              'target-arrow-color': '#ccc',
              'target-arrow-shape': 'triangle',
            },
          },
        ],
        layout: {
          name: "fcose"
        }
      }
    }
  }
}
</script>