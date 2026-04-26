window.APP_CONFIG = {
    githubLink: "https://github.com/kkonst40/cloud-storage",
    mainName: "CLOUD STORAGE",
    baseUrl: "http://localhost:8080",
    baseApi: "/api",

    validateLoginForm: true,
    validateRegistrationForm: true,

    validUsername: {
        minLength: 5,
        maxLength: 20,
        pattern: "^[a-zA-Z0-9]+[a-zA-Z_0-9]*[a-zA-Z0-9]+$",
    },

    validPassword: {
        minLength: 8,
        maxLength: 72,
        pattern: "^[a-zA-Z0-9!@#$%^&*(),.?\":{}|<>[\\]/`~+=-_';]*$",
    },

    validFolderName: {
        minLength: 1,
        maxLength: 200,
        pattern: "^[^/\\\\:*?\"<>|]+$",
    },

    //Разрешать ли перемещение выделенных файлов и папок с помощью перетаскивания в соседние папки. (drag n drop)
    isMoveAllowed: true,

    //Разрешить вырезать и вставлять файлы/папки.
    isCutPasteAllowed: true,

    //Разрешить кастомное контекстное меню для управления файлами (вызывается правой кнопкой мыши - на одном файле, или на выделенных)
    isFileContextMenuAllowed: true,

    //Разрешить шорткаты на странице - Ctrl+X, Ctrl+V, Del - на выделенных элементах
    isShortcutsAllowed: true,

    functions: {
        mapObjectToFrontFormat: (obj) => {
            return {
                lastModified: null,
                name: obj.name,
                size: obj.size,
                path: obj.path + obj.name,
                folder: obj.type === "DIRECTORY"
            }
        },

    }

};