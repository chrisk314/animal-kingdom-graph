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
    },
  },
  methods: {
    // on-click event listener for nodes
    getChildNodes: function(event) {
      const node = event.target;
      const nodeId = node.id();
      console.log("Clicked node: " + nodeId);
      console.log("Successors in graph: " + node.successors().map( (node) => node.id() ));

      // Check if child nodes are already displayed
      const childElems = node.successors();
      if (childElems.length > 0) {
        // Remove child nodes and edges from graph
        console.log("Removing successors from graph.");
        this.cy.remove(childElems);
        return;
      }

      // Check if child nodes have already been fetched
      if (this.childNodeCache[nodeId]) {
        console.log("Adding cached elements to graph.");
        if (this.childNodeCache[nodeId].nodes.length == 0) {
          console.log("No child nodes found.");
          return;
        }
        // Add child nodes and edges to graph
        this.cy.add(this.childNodeCache[nodeId].nodes);
        this.cy.add(this.childNodeCache[nodeId].edges);
        // Layout graph
        this.cy.layout(this.data.layout).run();
        return;
      }

      // Call backend api to get child nodes
      let childNodesUrl = `${this.backend_api_baseurl}/taxon/${nodeId}/children`;
      console.log("Retreiving nodes from backend API: " + childNodesUrl);
      fetch(childNodesUrl)
        .then(response => response.json())
        .then(data => {
          // Build child nodes data
          let childNodes = data.data.map( (child) => {
            return {
              group: "nodes",
              data: child,
            }
          })
          // Build child edges data
          let childEdges = data.data.map( (child) => {
            return {
              group: "edges",
              data: {
                id: `${nodeId}-${child.id}`,
                source: nodeId,
                target: child.id
              },
            }
          })
          // Cache child nodes and edges
          this.childNodeCache[nodeId] = { nodes: childNodes, edges: childEdges };
          if (childNodes.length == 0) {
            console.log("No child nodes found.");
            return;
          }
          // Add child nodes and edges to graph
          this.cy.add(childNodes);
          this.cy.add(childEdges);
          // Layout graph
          this.cy.layout(this.data.layout).run();
        })
        .catch(error => console.error(error));
    },
    showIframe(url) {
      const iframe = document.createElement('iframe')
      iframe.src = url
      iframe.style.position = 'absolute'
      iframe.style.top = '0'
      iframe.style.left = '0'
      iframe.style.width = '100%'
      iframe.style.height = '100%'
      iframe.style.border = 'none'
      document.body.appendChild(iframe)
      this.iframe = iframe
    },
    hideIframe() {
      if (this.iframe) {
        document.body.removeChild(this.iframe)
        this.iframe = null
      }
    }
  },
  mounted() {
    this.childNodeCache = {};
    this.backend_api_baseurl = import.meta.env.VITE_BACKEND_API_BASEURL;

    this.cy = cytoscape({
      container: this.$refs.cy,
      elements: this.data.elements,
      style: this.data.style,
      layout: this.data.layout
    })
    
    // Add event listener for node click
    this.cy.on('click', 'node', this.getChildNodes)
    
    // Add event listener for node hover
    this.cy.on('mouseover', 'node', (event) => {
      this.hoverTimeout = setTimeout(() => {
        const node = event.target
        const wikipediaUrl = node.data('url')
        if (wikipediaUrl) {
          this.showIframe(wikipediaUrl)
        }
      }, 1000)
    })

    // Add event listener for node unhover
    this.cy.on('mouseout', 'node', () => {
      clearTimeout(this.hoverTimeout)
      this.hideIframe()
    })
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