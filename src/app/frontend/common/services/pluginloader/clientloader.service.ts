import { Injectable, NgModuleFactory } from '@angular/core';
import { PluginLoaderService } from './pluginloader.service';
import { PLUGIN_EXTERNALS_MAP } from './pluginexternals';
import { PluginsConfigProvider } from './pluginsconfig.provider';

const systemJS = window.System;

@Injectable()
export class ClientPluginLoaderService extends PluginLoaderService {
  constructor(private configProvider: PluginsConfigProvider) {
    super();
  }

  provideExternals() {
    Object.keys(PLUGIN_EXTERNALS_MAP).forEach(externalKey =>
      window.define(externalKey, [], () => {
        // @ts-ignore
        return PLUGIN_EXTERNALS_MAP[externalKey];
      })
    );
  }

  load<T>(pluginName: string): Promise<NgModuleFactory<T>> {
    const { config } = this.configProvider;
    if (!config[pluginName]) {
      throw Error(`Can't find appropriate plugin`);
    }

    const depsPromises = (config[pluginName].deps || []).map(dep => {
      return systemJS.import(config[dep].path).then(m => {
        window['define'](dep, [], () => m.default);
      });
    });

    return Promise.all(depsPromises).then(() => {
      return systemJS
        .import(config[pluginName].path)
        .then(module => module.default.default);
    });
  }
}
