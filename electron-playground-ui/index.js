// const { app, BrowserWindow } = require('electron')
const { app, BrowserWindow, ipcMain, dialog, globalShortcut } = require('electron')
const Store = require('electron-store');
const store = new Store();

ipcMain.handle('save-ssh-params', (event, params) => {
  console.log("Saving parameters handle:", params);
  try {
      store.set('sshParams', params);
      console.log("Successfully saved:", params);
      return true;
  } catch (err) {
      console.error("Error saving parameters:", err);
      return false;
  }
});

ipcMain.handle('get-ssh-params', event => {
  return store.get('sshParams', {});
});

function createWindow () {
  const win = new BrowserWindow({
    width: 800,
    height: 600,
    webPreferences: {
        nodeIntegration: true, // keep this as false for security
        contextIsolation: true, // protect against prototype pollution attacks
        enableRemoteModule: false, // turn off remote
        preload: __dirname + '/preload.js' // use the preload script
    }
  })

  win.loadFile('index.html')
  win.webContents.openDevTools();
}

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow()
  }
})

ipcMain.handle('open-file-dialog', async (event) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    const options = {
        title: 'Select Private Key',
        defaultPath: '~/.ssh', // start in the .ssh directory
        // defaultPath: '/', // start in the .ssh directory
        buttonLabel: 'Select',
        properties: ['openFile'],
        filters: [
            { name: 'All Files', extensions: ['*'] }
        ]
    };

    const result = await dialog.showOpenDialog(window, options);

    if (result.canceled) {
        return '';
    } else {
        return result.filePaths[0]; // return the selected path
    }
});

app.on('ready', () => {
  // Intercept and prevent the default shortcut for opening dev tools
  globalShortcut.register('CommandOrControl+Shift+I', () => {
      console.log('Intercepted CommandOrControl+Shift+I');
  });

  globalShortcut.register('CommandOrControl+Alt+I', () => {
      console.log('Intercepted CommandOrControl+Alt+I');
  });

  // ... rest of your code (like creating the BrowserWindow, etc.)
});