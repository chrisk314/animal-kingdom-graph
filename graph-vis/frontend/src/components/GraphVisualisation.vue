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