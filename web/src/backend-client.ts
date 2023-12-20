import { cancellableFetch } from './cancellable-fetch';

const BACKEND_ENDPOINT = '/api/proxy/plugin/observability-ui-hub/backend';

export interface UIPluginResponse {
  name: string;
  displayName: string;
  version: string;
  type: string;
  isEnabled: boolean;
  provider?: string;
  description?: string;
  iconImg?: string;
}

const getCSRFToken = () => {
  const cookiePrefix = 'csrf-token=';
  return (
    (document &&
      document.cookie &&
      document.cookie
        .split(';')
        .map((c) => c.trim())
        .filter((c) => c.startsWith(cookiePrefix))
        .map((c) => c.slice(cookiePrefix.length))
        .pop()) ??
    ''
  );
};

export const getPluginRequest = async (pluginName: string) => {
  return cancellableFetch<UIPluginResponse>(`${BACKEND_ENDPOINT}/api/v1/plugins/${pluginName}`);
};

export const listPluginsRequest = async () => {
  return cancellableFetch<Array<UIPluginResponse>>(`${BACKEND_ENDPOINT}/api/v1/plugins`);
};

export const enablePluginRequest = async (pluginType: string) => {
  return cancellableFetch<Array<UIPluginResponse>>(`${BACKEND_ENDPOINT}/api/v1/plugins/enable`, {
    method: 'POST',
    body: JSON.stringify({ type: pluginType }),
    headers: {
      'Content-Type': 'application/json',
      'X-CSRFToken': getCSRFToken(),
    },
  });
};

export const deletePluginRequest = async (pluginName: string) => {
  return cancellableFetch<Array<UIPluginResponse>>(
    `${BACKEND_ENDPOINT}/api/v1/plugins/${pluginName}`,
    {
      method: 'DELETE',
      headers: {
        'X-CSRFToken': getCSRFToken(),
      },
    },
  );
};
