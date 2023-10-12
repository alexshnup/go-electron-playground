const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld(
  "api", {
      invoke: (channel, ...args) => {
          return ipcRenderer.invoke(channel, ...args);
      }
  }
);




