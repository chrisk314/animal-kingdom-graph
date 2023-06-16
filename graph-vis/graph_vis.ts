import cytoscape from 'cytoscape';
import { Database } from "arangojs/database";

// Store taxonomic heirarchy in map
const taxonomicRanks = new Map([
    ['kingdom', 'phylum'],
    ['phylum', 'class'],
    ['class', 'order'],
    ['order', 'family'],
    ['family', 'genus'],
    ['genus', 'species'],
]);

// Connect to ArangoDB
const db = new Database({
  url: 'http://localhost:8529',
  databaseName: 'animal_kingdom',
  auth: { username: 'root', password: 'password' },
});

// Define the Cytoscape.js graph
const cy = cytoscape({
  container: document.getElementById('cy'),
  layout: { name: 'grid' },
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
});

// Load the root node of the animal kingdom graph
let rootNode: any;
db.query(`FOR t IN kingdom FILTER t.name == 'Animalia' RETURN t`)
  .then((cursor) => cursor.next())
  .then((taxon) => {
    rootNode = taxon;
    cy.add({ data: { id: rootNode._id, name: rootNode.name, rank: rootNode.rank } });
  })
  .catch((err) => console.error(err));

// Add click event listener to nodes
cy.on('click', 'node', (evt) => {
  const node = evt.target;
  const nodeId = node.id();
  const nodeRank = node.data('rank');
  const taxCollName = taxonomicRanks.get(nodeRank);
  const edgeCollName = `${nodeRank}Members`;

  // Load the child nodes of the clicked node
  db.query(
    `FOR t IN ${taxCollName} FILTER t._id IN (FOR e IN ${edgeCollName} FILTER e._to == '${nodeId}' RETURN e._from) RETURN t`
  )
    .then((cursor) => {
      const childNodes = [];
      cursor.forEach((taxon) => {
        childNodes.push({ data: { id: taxon._id, name: taxon.name, rank: taxon.rank } });
        cy.add({ data: { id: taxon._id, name: taxon.name, rank: taxon.rank } });
        cy.add({ data: { source: nodeId, target: taxon._id } });
      });
      node.children(childNodes);  // TODO : What is the purpose of adding children to the node?
    })
    .catch((err) => console.error(err));
});

// Add mouseover event listener to nodes
cy.on('mouseover', 'node', (evt) => {
  const node = evt.target;
  const nodeId = node.id();
  const taxCollName = node.data('rank');
  const taxColl = db.collection(taxCollName);

  // Load the summary of the Wikipedia page for the node
  taxColl.document(nodeId).then((taxon) => {
    // const summary = taxon.summary;
    // const image = taxon.image;
    // node.qtip({
    //   content: {
    //     title: node.data('name'),
    //     text: `<img src="${image}" /><br>${summary}`,
    //   },
    //   show: { event: evt.type, ready: true },
    //   hide: { event: 'mouseout unfocus' },
    //   style: { classes: 'qtip-bootstrap', tip: { width: 16, height: 8 } },
    // });
    node.qtip({
      content: {
        title: node.data('name'),
        text: `<h2>${node.data('name')}</h2>`,
      },
      show: { event: evt.type, ready: true },
      hide: { event: 'mouseout unfocus' },
      style: { classes: 'qtip-bootstrap', tip: { width: 16, height: 8 } },
    });
  });
});

// Add mouseout event listener to nodes
cy.on('mouseout', 'node', (evt) => {
  const node = evt.target;
  node.qtip('destroy');
});