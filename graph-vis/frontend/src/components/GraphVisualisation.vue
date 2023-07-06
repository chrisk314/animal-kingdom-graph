<template>
  <div id="cy" ref="cy"></div>
</template>

<script>
import cytoscape from 'cytoscape'
import fcose from 'cytoscape-fcose';

cytoscape.use( fcose );

export default {
  name: 'GraphVisualisation',
  props: {
    data: {
      type: Object,
      required: true
    }
  },
  methods: {
    // on-click event listener for nodes
    getChildNodes: function(event) {
      const node = event.target;
      const nodeId = node.id();

      console.log("Clicked node: " + nodeId);

      // Call backend api to get child nodes
      fetch(`http://localhost:5000/api/nodes/${nodeId}/children`)
        .then(response => response.json())
        .then(data => {
          // Add child nodes to graph
          this.cy.add(data);
          // Add edges between parent and child nodes
          this.cy.add({
            data: {
              id: `${nodeId}-child`,
              source: nodeId,
              target: data[0].data.id
            }
          });
          // Layout graph
          this.cy.layout(this.data.layout).run();
        })
        .catch(error => console.error(error));
    }
  },
  mounted() {
    this.cy = cytoscape({
      container: this.$refs.cy,
      elements: this.data.elements,
      style: this.data.style,
      layout: this.data.layout
    })
    this.cy.on('click', 'node', this.getChildNodes)
  },
  beforeDestroy() {
    this.cy.destroy()
  },
}
</script>

<style scoped>
/* TODO : How to select ref="cy" element with CSS selectors? */
#cy {
    /* TODO : Figure out scroll bar popping in and out with 100% */
  /* height: 100%;
  width: 100%; */
  height: 99%;
  width: 99%;
  position: absolute;
  top: 0px;
  left: 0px;
}
</style>