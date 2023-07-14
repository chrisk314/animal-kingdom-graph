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
          { group: "nodes", data: { id: "kingdom/80455", rank: "Kingdom", name: "Animalia", url: "https://en.wikipedia.org/wiki/Animal" } },
        ],
        style: [
          {
            selector: 'node',
            style: {
              'background-color': (node) => {
                // Set node colour based on rank
                switch (node.data('rank')) {
                  case 'Kingdom':
                    return 'red';
                  case 'Phylum':
                    return 'orange';
                  case 'Class':
                    return 'yellow';
                  case 'Order':
                    return 'green';
                  case 'Family':
                    return 'blue';
                  case 'Genus':
                    return 'purple';
                  case 'Species':
                    return 'pink';
                  default:
                    return 'grey';
                }
              },
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