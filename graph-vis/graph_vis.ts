import cytoscape from 'cytoscape';
import arangojs, { Database } from 'arangojs';

// Connect to ArangoDB
const db = new Database({
  url: 'http://localhost:8529',
  databaseName: 'mydb',
  auth: { username: 'myuser', password: 'mypassword' },
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
db.query(`FOR t IN Taxa FILTER t.name == 'Animalia' RETURN t`)
  .then((cursor) => cursor.next())
  .then((taxon) => {
    rootNode = taxon;
    cy.add({ data: { id: rootNode._id, name: rootNode.name } });
  })
  .catch((err) => console.error(err));

// Add click event listener to nodes
cy.on('click', 'node', (evt) => {
  const node = evt.target;
  const nodeId = node.id();
  const nodeRank = node.data('rank');
  const edgeCollName = `${nodeRank}Members`;
  const taxCollName = `${nodeRank}Taxa`;
  const taxColl = db.collection(taxCollName);
  const edgeColl = db.edgeCollection(edgeCollName);

  // Load the child nodes of the clicked node
  db.query(
    `FOR t IN Taxa FILTER t._id IN (FOR e IN ${edgeCollName} FILTER e._from == '${nodeId}' RETURN e._to) RETURN t`
  )
    .then((cursor) => {
      const childNodes = [];
      cursor.each((taxon) => {
        childNodes.push({ data: { id: taxon._id, name: taxon.name, rank: taxon.rank } });
        cy.add({ data: { id: taxon._id, name: taxon.name, rank: taxon.rank } });
        cy.add({ data: { source: nodeId, target: taxon._id } });
      });
      node.children(childNodes);
    })
    .catch((err) => console.error(err));
});

// Add mouseover event listener to nodes
cy.on('mouseover', 'node', (evt) => {
  const node = evt.target;
  const nodeId = node.id();
  const taxCollName = `${node.data('rank')}Taxa`;
  const taxColl = db.collection(taxCollName);

  // Load the summary of the Wikipedia page for the node
  taxColl.document(nodeId).then((taxon) => {
    const summary = taxon.summary;
    const image = taxon.image;
    node.qtip({
      content: {
        title: node.data('name'),
        text: `<img src="${image}" /><br>${summary}`,
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