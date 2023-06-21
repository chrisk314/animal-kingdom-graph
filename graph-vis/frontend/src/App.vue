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
          // graph elements here
        ],
        style: [
          // graph style here
        ],
        layout: {
          // graph layout here
        }
      }
    }
  }
}
</script>