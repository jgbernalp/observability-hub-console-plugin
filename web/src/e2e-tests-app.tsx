import '@patternfly/patternfly/patternfly.css';
import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter, Link, Route } from 'react-router-dom';
import i18n from './i18n';
import './index.css';
import CatalogPage from './pages/catalog-page';

const EndToEndTestsApp = () => {
  return (
    <div className="pf-c-page co-logs-standalone__page">
      <BrowserRouter>
        <header className="pf-c-masthead">
          <div className="pf-c-masthead__main"></div>
        </header>

        <div className="pf-c-page__sidebar co-logs-standalone__side-menu">
          <div className="pf-c-page__sidebar-body">
            <nav className="pf-c-nav" aria-label="Global">
              <ul className="pf-c-nav__list">
                <li className="pf-c-nav__item">
                  <Link className="pf-c-nav__link" to="/observability-ui/catalog">
                    Catalog
                  </Link>
                </li>
              </ul>
            </nav>
          </div>
        </div>

        <main className="pf-c-page__main" tabIndex={-1}>
          <Route path="/observability-ui/catalog">
            <CatalogPage />
          </Route>
        </main>
      </BrowserRouter>
    </div>
  );
};

i18n.on('initialized', () => {
  ReactDOM.render(<EndToEndTestsApp />, document.getElementById('app'));
});
