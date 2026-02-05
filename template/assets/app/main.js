function app() {
            return {
                currentTab: 'server',
                lang: localStorage.getItem('pz_lang') || 'CN', // 记住用户选择
                i18n: {}, // 存放当前语言的 UI 文本
                languageList: [], // 存放从后端获取的语言列表
                loading: false,
                serverConfig: [],
                sandboxConfig: [],
                serverSections: {},
                sandboxSections: {},
                logs: '...',
                status: { pid: 0, uptime: '0s' },
                toast: { show: false, message: '', type: 'success' },
                modInput: '',
                modLoading: false,
                availableMods: [], // 本地库
                activeMods: [],    // 当前启用列表
                logConnected: false,


                init() {
                    this.refreshAll();
                    // 监听 Tab 切换，触发sse
                    this.$watch('currentTab', (val) => {
                        if (val === 'monitor') {
                            this.startLogStream();
                        } else {
                            this.stopLogStream(); // 离开页面时断开连接，节省资源
                        }
                    });
                },

                refreshAll() {
                    this.fetchConfig('server');
                    this.fetchConfig('sandbox');
                    // 预留: this.fetchStatus();
                },

                async initI18n() {
                    this.loading = true;
                    try {
                        // 请求语言列表和 UI 文本
                        const res = await fetch(`/api/i18n?lang=${this.lang}`);
                        const data = await res.json();
                        
                        this.languageList = data.languages;
                        this.i18n = data.ui;
                        this.logs = this.i18n.log_refresh_hint || 'Click refresh...';
                        
                        // 加载完 I18n 后再加载配置
                        this.refreshAll();
                    } catch (e) {
                        console.error("Failed to init i18n", e);
                    } finally {
                        this.loading = false;
                    }
                },

                async switchLanguage() {
                    localStorage.setItem('pz_lang', this.lang);
                    // 重新加载 UI 文本和配置项翻译
                    await this.initI18n();
                },

                // 统一获取配置并分组
                async fetchConfig(type) {
                    this.loading = true;
                    try {
                        const res = await fetch(`/api/config/${type}?lang=${this.lang}`);
                        const data = await res.json();
                        
                        // 按 Section 分组
                        const grouped = data.items.reduce((acc, item) => {
                            const section = item.section || 'General';
                            if (!acc[section]) acc[section] = [];
                            acc[section].push(item);
                            return acc;
                        }, {});

                        if (type === 'server') {
                            this.serverConfig = data.items; // 保存原始数组用于提交
                            this.serverSections = grouped;
                        } else {
                            this.sandboxConfig = data.items;
                            this.sandboxSections = grouped;
                        }
                    } catch (e) {
                        this.showToast((this.i18n.msg_config_load_fail || 'Load failed') + ': ' + e, 'error');
                    } finally {
                        this.loading = false;
                    }
                },

                // 打开管理器
                async openModManager() {
                    this.modLoading = true;
                    const availableModsResp = await fetch(`/api/mods`);
                    this.availableMods = await availableModsResp.json()
                    // 从 serverConfig 解析当前配置
                    const modsItem = this.serverConfig.find(i => i.key === 'Mods');
                    const wsItem = this.serverConfig.find(i => i.key === 'WorkshopItems');
                    
                    const currentModIds = (modsItem ? modsItem.value : "")
                    .split(';')
                    .map(s => s.trim().replace(/^\\/, '')) // 去除开头的 \
                    .filter(Boolean);
                    const currentWsIds = (wsItem ? wsItem.value : "").split(';').filter(Boolean);
                    // 如果没有任何工坊物品，直接清空列表并打开
                    if (currentWsIds.length === 0) {
                        this.modLoading = false;
                        this.activeMods = [];
                        document.getElementById('mod_modal').showModal();
                        return;
                    }
                    document.getElementById('mod_modal').showModal();
                    const res = await fetch(`/api/mods/lookup?ids=${currentWsIds.join(',')}`);
                    const lookupData = await res.json();
                    this.modLoading = false;
                    this.activeMods = [];
                    currentWsIds.forEach(wid => {
                        const info = lookupData.find(d => d.workshop_id === wid);
                        if (info) {
                            // 一个 Workshop Item 可能包含多个 Mod ID
                            // 需要判断 currentModIds 里包含 info.mod_id 里的哪些
                            // 获取该工坊物品包含的所有潜在 ModID (后端返回的是逗号分隔字符串 "id1,id2")
                            const potentialModIds = info.mod_id.split(',').map(s => s.trim());
                            
                            // 在当前启用的 Mods 里找，有哪些是属于这个 Workshop 的
                            const enabledSubMods = potentialModIds.filter(pmid => currentModIds.includes(pmid));
                            
                            if (enabledSubMods.length > 0) {
                                // 如果找到了匹配的，添加进列表
                                enabledSubMods.forEach(mid => {
                                    this.activeMods.push({
                                        name: info.name + (enabledSubMods.length > 1 ? ` (${mid})` : ''),
                                        workshop_id: wid,
                                        mod_id: mid
                                    });
                                });
                            }
                        } else {
                            // 没查到信息 (网络错误或ID不存在)，显示 ID
                            this.activeMods.push({
                                name: `Unknown Item (${wid})`,
                                workshop_id: wid,
                                mod_id: '?'
                            });
                        }
                    });
                    document.getElementById('mod_modal').showModal();
                },

                // 解析输入框并添加
                async lookupAndAddMods() {
                    if (!this.modInput.trim()) return;
                    
                    // 处理逗号分隔
                    const rawIds = this.modInput.split(/[,，;\n]/).map(s => s.trim()).filter(Boolean);
                    if (rawIds.length === 0) return;

                    this.modLoading = true;
                    try {
                        const res = await fetch(`/api/mods/lookup?ids=${rawIds.join(',')}`);
                        const data = await res.json();
                        
                        for (const item of data) {
                            this.addModItem(item);
                        }
                        
                        this.modInput = ''; // 清空输入框
                    } catch (e) {
                         this.showToast((this.i18n.msg_parse_fail || 'Parse failed') + ': ' + e, 'error');
                    } finally {
                        this.modLoading = false;
                    }
                },

                // 添加单个 Mod 项目 (处理多 ModID 情况)
                addModItem(item) {
                    // 检查 ModID 是否存在
                    let mid = item.mod_id;
                    
                    // 如果包含多个 ID (例如 "ID1,ID2")
                    if (mid.includes(',')) {
                        const choices = mid.split(',');
                        // 全部添加进去，最省事
                        choices.forEach(subId => {
                            this.pushToActive({
                                name: item.name + ` (${subId})`, // 区分名字
                                workshop_id: item.workshop_id,
                                mod_id: subId.trim()
                            });
                        });
                        return;
                    }

                    //  如果 ModID 未知 (?)
                    if (mid === '?' || mid === 'Unknown (Check Page)') {
                         const msg = (this.i18n.prompt_mod_id_manual || '').replace('{0}', item.name);
                        mid = prompt(msg || `Please enter Mod ID for ${item.name}:`);
                        if (!mid) return;
                    }
                    
                    this.pushToActive({
                        name: item.name,
                        workshop_id: item.workshop_id,
                        mod_id: mid
                    });
                },
                //添加到列表并去重
                pushToActive(modObj) {
                    // 检查是否已存在 (根据 ModID)
                    if (this.activeMods.some(m => m.mod_id === modObj.mod_id)) return;
                    this.activeMods.push(modObj);
                },
                // 从本地库添加
                addFromLocal(mod) {
                    this.pushToActive(mod);
                },

                // 移除
                removeMod(index) {
                    this.activeMods.splice(index, 1);
                },

                // 排序
                moveMod(index, delta) {
                    const newIndex = index + delta;
                    if (newIndex < 0 || newIndex >= this.activeMods.length) return;
                    
                    const temp = this.activeMods[index];
                    this.activeMods[index] = this.activeMods[newIndex];
                    this.activeMods[newIndex] = temp;
                },

                // 保存回 Server Config
                saveModsToConfig() {
                    // 提取 Mod IDs (分号分隔)
                    const modsStr = this.activeMods
                    .map(m => `\\${m.mod_id}`) // 加反斜杠，配置是那么写的。
                    .join(';');
                    
                    // 提取 Workshop IDs (分号分隔，去重)
                    // 注意：过滤掉 '?' 和本地已有的 workshopID
                    const wsIds = [...new Set(this.activeMods.map(m => m.workshop_id).filter(id => id && id !== '?'))];
                    const wsStr = wsIds.join(';');

                    // 更新 serverConfig 数组
                    const modsItem = this.serverConfig.find(i => i.key === 'Mods');
                    const wsItem = this.serverConfig.find(i => i.key === 'WorkshopItems');

                    if (modsItem) modsItem.value = modsStr;
                    if (wsItem) wsItem.value = wsStr;

                    document.getElementById('mod_modal').close();
                    this.showToast(this.i18n.msg_mod_list_updated, 'success');
                },

                // 保存配置
                async saveConfig(type, restart) {
                    if (restart && !confirm(this.i18n.confirm_save_restart)) return;
                    this.loading = true;
                    try {
                        const items = type === 'server' ? this.serverConfig : this.sandboxConfig;

                        const res = await fetch(`/api/config/${type}`, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ items: items, restart: restart })
                        });

                        if (!res.ok) throw new Error(this.i18n.msg_save_fail);

                        this.showToast(this.i18n.msg_save_success, 'success');

                    } catch (e) {
                        this.showToast(e.message, 'error');
                    } finally {
                        this.loading = false;
                    }
                },

                async performAction(action) {
                    this.loading = true;
                    try {
                        const res = await fetch(`/api/action/${action}`, { method: 'POST' });
                        if (res.ok) {
                             this.showToast((this.i18n.msg_cmd_sent || 'Command Sent') + ': ' + action, 'success');
                        } else {
                            throw new Error(this.i18n.msg_exec_fail);
                        }
                    } catch (e) {
                        this.showToast(e.message, 'error');
                    } finally {
                        this.loading = false;
                    }
                },

                 // 开启 SSE 日志流
                startLogStream() {
                    if (this.eventSource) return; // 避免重复连接

                    this.logs = "Connecting...\n";
                    
                    // 创建 EventSource
                    this.eventSource = new EventSource('/api/logs/stream');

                    this.eventSource.onopen = () => {
                        this.logConnected = true;
                        this.logs = ""; // 连接成功后清空提示
                    };

                    this.eventSource.onmessage = (event) => {
                        // 追加日志
                        this.logs += event.data + "\n";
                        
                        // 限制日志长度，防止浏览器内存爆炸（例如保留最近 5000 字符）
                        if (this.logs.length > 10000) {
                            this.logs = this.logs.slice(-5000);
                        }

                        // 自动滚动到底部
                        this.$nextTick(() => {
                            if (this.$refs.logBox) {
                                this.$refs.logBox.scrollTop = this.$refs.logBox.scrollHeight;
                            }
                        });
                    };

                    this.eventSource.onerror = (err) => {
                        console.error("SSE Error:", err);
                        this.logConnected = false;
                        this.eventSource.close();
                        this.eventSource = null;
                        // 可以选择自动重连，或者手动重连
                        this.logs += "\n[disconnect...]\n";
                    };
                },

                // 关闭 SSE 日志流
                stopLogStream() {
                    if (this.eventSource) {
                        this.eventSource.close();
                        this.eventSource = null;
                        this.logConnected = false;
                    }
                },
                // 切换日志流开关
                toggleLogStream() {
                    if (this.logConnected) {
                        this.stopLogStream();
                    } else {
                        this.startLogStream();
                    }
                },

                showToast(msg, type = 'success') {
                    this.toast.message = msg;
                    this.toast.type = type;
                    this.toast.show = true;
                    setTimeout(() => this.toast.show = false, 3000);
                },
                // 重启面板 API 调用
                restartPanel() {
                    if (!confirm(this.i18n.restart_in_progress)) return;

                    fetch('/api/service/restart', { method: 'POST' })
                        .then(res => res.json())
                        .then(data => {
                            alert(data.message);
                            // 3秒后刷新页面
                            setTimeout(() => location.reload(), 3000);
                        })
                        .catch(err => alert(this.i18n.restart_failed + err));
                },
                async checkUpdate() {
                    this.loading = true;
                    try {
                        const res = await fetch('/api/system/check_update');
                        const data = await res.json();
                        if (data.new_version) {
                            // 获取模板文本
                            let msg = this.i18n.msg_update_found || 'New version {0} found (Current: {1})\nUpdate now?';
                            
                            // 简单的手动替换占位符
                            msg = msg.replace('{0}', data.new_version)
                                    .replace('{1}', data.current);

                            if (confirm(msg)) {
                                await fetch('/api/system/perform_update', {
                                    method: 'POST',
                                    headers: {'Content-Type': 'application/json'},
                                    body: JSON.stringify({ url: data.download_url })
                                });
                                // 使用 i18n
                                alert(this.i18n.msg_update_performing || 'Update command sent, please refresh later.');
                            }
                        } else {
                            if(data.error) {
                                this.showToast(data.error, 'error');
                                return
                            }
                            // 使用 i18n
                            this.showToast(this.i18n.msg_already_latest || 'Already latest version');
                        }
                    } catch(e) {
                        // 使用 i18n
                        this.showToast(this.i18n.msg_update_check_fail || 'Check update failed', 'error');
                    } finally {
                        this.loading = false;
                    }
                }
            }
        }
