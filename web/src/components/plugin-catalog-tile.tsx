import { CatalogTile } from '@patternfly/react-catalog-view-extension';
import { Alert, Button, Grid, GridItem } from '@patternfly/react-core';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { UIPluginResponse } from '../backend-client';
import { isFetchError } from '../cancellable-fetch';
import { useUIPlugins } from '../hooks/usePlugins';

interface PluginCatalogTileProps {
  plugin: UIPluginResponse;
}

const isError = (error: unknown): error is Error => {
  return typeof error === 'object' && error !== null && 'message' in error;
};

export const PluginCatalogTile: React.FC<PluginCatalogTileProps> = ({ plugin }) => {
  const { enablePlugin, deletePlugin, isEnabling, getPlugin } = useUIPlugins();
  const [errorMessage, setErrorMessage] = React.useState<string | undefined>(undefined);
  const { t } = useTranslation('plugin__observability-ui-hub');

  const handleEnableClick = async () => {
    try {
      if (plugin.isEnabled) {
        await deletePlugin(plugin.name);
      } else {
        await enablePlugin(plugin.type);
      }
    } catch (error) {
      if (isError(error)) {
        setErrorMessage(error.message);
      } else {
        setErrorMessage('Unknown error');
      }
    }

    getPlugin(plugin.name).catch((err) => {
      if (!isFetchError(err) || err.status !== 404) {
        // eslint-disable-next-line no-console
        console.error(err);
      }
    });
  };

  return (
    <CatalogTile
      iconImg={plugin.iconImg}
      title={plugin.displayName}
      vendor={plugin.provider}
      description={plugin.description}
    >
      <Grid hasGutter>
        <GridItem>
          <Button
            isDisabled={isEnabling}
            isLoading={isEnabling}
            variant="primary"
            onClick={handleEnableClick}
          >
            {plugin.isEnabled ? t('Disable') : t('Enable')}
            {errorMessage && <Alert variant="danger" title={errorMessage} />}
          </Button>
        </GridItem>
      </Grid>
    </CatalogTile>
  );
};
