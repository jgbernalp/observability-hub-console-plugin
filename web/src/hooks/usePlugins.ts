import React from 'react';
import {
  UIPluginResponse,
  listPluginsRequest,
  enablePluginRequest,
  deletePluginRequest,
  getPluginRequest,
} from '../backend-client';

const isAbortError = (error: unknown): boolean =>
  error instanceof Error && error.name === 'AbortError';

export const useUIPlugins = () => {
  const [plugins, setPlugins] = React.useState<Array<UIPluginResponse>>([]);
  const [isEnabling, setIsEnabling] = React.useState<boolean>(false);
  const listPluginsAbort = React.useRef<() => void | undefined>();
  const togglePluginAbort = React.useRef<() => void | undefined>();

  const listPlugins = async () => {
    if (listPluginsAbort.current) {
      listPluginsAbort.current();
    }
    try {
      const { abort, request } = await listPluginsRequest();

      listPluginsAbort.current = abort;

      const pluginsList = await request();
      setPlugins(pluginsList);
    } catch (error) {
      if (!isAbortError(error)) {
        throw error;
      }
    }
  };

  const getPlugin = async (pluginName: string) => {
    if (listPluginsAbort.current) {
      listPluginsAbort.current();
    }
    try {
      const { abort, request } = await getPluginRequest(pluginName);

      listPluginsAbort.current = abort;

      const pluginItem = await request();

      setPlugins(plugins.map((plugin) => (plugin.name === pluginItem.name ? pluginItem : plugin)));
    } catch (error) {
      if (!isAbortError(error)) {
        throw error;
      }
    }
  };

  const enablePlugin = async (pluginType: string) => {
    setIsEnabling(true);
    if (togglePluginAbort.current) {
      togglePluginAbort.current();
    }
    try {
      const { abort, request } = await enablePluginRequest(pluginType);

      togglePluginAbort.current = abort;

      await request();
    } catch (error) {
      if (!isAbortError(error)) {
        throw error;
      }
    } finally {
      setIsEnabling(false);
    }
  };

  const deletePlugin = async (pluginName: string) => {
    setIsEnabling(true);

    if (togglePluginAbort.current) {
      togglePluginAbort.current();
    }
    try {
      const { abort, request } = await deletePluginRequest(pluginName);

      togglePluginAbort.current = abort;

      await request();
    } catch (error) {
      if (!isAbortError(error)) {
        throw error;
      }
    } finally {
      setIsEnabling(false);
    }
  };

  return {
    listPlugins,
    plugins,
    getPlugin,
    enablePlugin,
    deletePlugin,
    isEnabling,
  };
};
