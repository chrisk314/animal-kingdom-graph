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
        elements: [
          { data: { id: 'a', name: 'Node A' } },
          { data: { id: 'b', name: 'Node B' } },
          { data: { id: 'c', name: 'Node C' } },
          { data: { id: 'd', name: 'Node D' } },
          { data: { id: 'e', name: 'Node E' } },
          { data: { id: 'f', name: 'Node F' } },
          { data: { id: 'ab', source: 'a', target: 'b' } },
          { data: { id: 'ac', source: 'a', target: 'c' } },
          { data: { id: 'bc', source: 'b', target: 'c' } },
          { data: { id: 'cd', source: 'c', target: 'd' } },
          { data: { id: 'de', source: 'd', target: 'e' } },
          { data: { id: 'ef', source: 'e', target: 'f' } },
        ],
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