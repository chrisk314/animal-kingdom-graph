<template>
  <div ref="cy"></div>
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
  mounted() {
    this.cy = cytoscape({
      container: this.$refs.cy,
      elements: this.data.elements,
      style: this.data.style,
      layout: this.data.layout
    })
  },
  beforeDestroy() {
    this.cy.destroy()
  }
}
</script>

<style scoped>
div {
  height: 100%;
  width: 100%;
}
</style>