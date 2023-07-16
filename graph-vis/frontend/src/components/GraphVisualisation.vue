<template>
  <div id="cy" ref="cy"></div>
  <div class="iframe-container" v-if="IframeIsVisible" @mouseenter="preventHideIframe" @mouseleave="hideIframe">
    <div class="iframe-shadow"></div>
    <iframe :src="iframeUrl" class="iframe"></iframe>
  </div>
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
    showIframe(event) {
      if (this.hideIframeTimeout) {
        clearTimeout(this.hideIframeTimeout);
        this.hideIframeTimeout = null;
      }
      if (this.showIframeTimeout) {
        clearTimeout(this.showIframeTimeout);
      }
      this.showIframeTimeout = setTimeout(() => {
        const node = event.target
        const wikipediaUrl = node.data('url')
        if (wikipediaUrl) {
          this.IframeIsVisible = true
          this.iframeUrl = wikipediaUrl
        }
        this.showIframeTimeout = null;
      }, 1000);
    },
    hideIframe(event) {
      clearTimeout(this.showIframeTimeout)
      this.showIframeTimeout = null;
      if (this.hideIframeTimeout) {
        return;
      }
      if (!this.IframeIsVisible) {
        return;
      }
      this.hideIframeTimeout = setTimeout(() => {
        this.IframeIsVisible = false
        this.iframeUrl = ''
        this.hideIframeTimeout = null;
      }, 500);
    },
    preventHideIframe() {
      clearTimeout(this.hideIframeTimeout)
      this.hideIframeTimeout = null;
    }
  },
  data() {
    return {
      IframeIsVisible: false,
      iframeUrl: ''
    }
  },
  mounted() {
    this.childNodeCache = {};
    this.backend_api_baseurl = import.meta.env.VITE_BACKEND_API_BASEURL;
    this.hideIframeTimeout = null;
    this.showIframeTimeout = null;

    this.cy = cytoscape({
      container: this.$refs.cy,
      elements: this.data.elements,
      style: this.data.style,
      layout: this.data.layout
    })
    
    // Add event listener for node click
    this.cy.on('click', 'node', this.getChildNodes)
    
    // Add event listener for node hover
    this.cy.on('mouseover', 'node', this.showIframe)

    // Add event listener for node unhover
    this.cy.on('mouseout', 'node', this.hideIframe)

    // Add event listener for node drag
    this.cy.on('grabon', 'node', this.hideIframe)
    
    // Add event listener for node drag end
    this.cy.on('dragfree', 'node', this.showIframe)
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
.iframe-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 1000;
  width: 80%;
  max-width: 800px;
  height: 80%;
  max-height: 600px;
  background-color: white;
  box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.iframe-shadow {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: black;
  opacity: 0.5;
}

.iframe {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border: none;
}
</style>