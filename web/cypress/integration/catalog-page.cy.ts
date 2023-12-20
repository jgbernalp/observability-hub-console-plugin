describe('Catalog Page', () => {
  it('renders correctly with an expected response', () => {
    cy.visit('observability-ui/catalog');

    cy.contains('Observability UI Plugin Catalog');
  });
});
