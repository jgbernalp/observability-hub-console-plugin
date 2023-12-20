import { Grid, GridItem, PageSection, Title } from '@patternfly/react-core';
import React, { useEffect } from 'react';
import { Helmet } from 'react-helmet';
import { useTranslation } from 'react-i18next';
import { PluginCatalogTile } from '../components/plugin-catalog-tile';
import { useUIPlugins } from '../hooks/usePlugins';

const CatalogPage: React.FC = () => {
  const { plugins, listPlugins: getPlugins } = useUIPlugins();
  const { t } = useTranslation('plugin__observability-ui-hub');

  useEffect(() => {
    getPlugins();
  }, []);

  return (
    <>
      <Helmet>
        <title>{t('Observability UI Plugin Catalog')}</title>
      </Helmet>
      <PageSection>
        <Grid hasGutter>
          <Title headingLevel="h1" size="lg">
            {t('Observability UI Plugin Catalog')}
          </Title>
          <Grid hasGutter>
            {plugins.map((plugin) => (
              <GridItem span={4} key={`${plugin.name}-${plugin.version}`}>
                <PluginCatalogTile plugin={plugin} />
              </GridItem>
            ))}
          </Grid>
        </Grid>
      </PageSection>
    </>
  );
};

export default CatalogPage;
