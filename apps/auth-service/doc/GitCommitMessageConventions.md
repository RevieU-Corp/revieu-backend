# 🧭 Git Commit Message Conventions / Git 提交信息规范

To maintain clear project history, all commit messages should follow the structure below:
为了保持项目提交历史的清晰性，所有提交信息应遵循以下结构：

```
<type>(<scope>): <subject>

<body>

<footer>
```

---

## 🧩 1. Type (类型，必填)

Describes the purpose of this change.
说明本次提交的目的或类型。

| Type         | Meaning / 含义                                               |
| ------------ | ---------------------------------------------------------- |
| **feat**     | ✨ A new feature / 新功能                                      |
| **fix**      | 🐛 A bug fix / 修复 Bug                                      |
| **docs**     | 📝 Documentation only changes / 仅修改文档                      |
| **style**    | 💅 Code style changes (no logic change) / 代码格式调整（无逻辑变化）    |
| **refactor** | 🔧 Code refactoring (no new feature or fix) / 重构（既非新增也非修复） |
| **perf**     | ⚡ Performance improvements / 性能优化                          |
| **test**     | ✅ Add or modify tests / 添加或修改测试                            |
| **build**    | 🏗️ Build process or dependency changes / 构建系统或依赖更新        |
| **ci**       | 🤖 CI/CD configuration changes / 持续集成配置修改                  |
| **chore**    | 🔩 Other minor changes / 杂项（不影响源码或测试）                      |
| **revert**   | ⏪ Revert a previous commit / 回滚先前提交                        |

---

## 🏷️ 2. Scope (影响范围，可选)

Specifies the module, file, or feature that is affected.
用于说明本次更改影响的模块或范围。

Examples / 示例：

```
feat(api): add user authentication
fix(ui): correct button alignment
refactor(core): optimize data processing logic
```

---

## 📝 3. Subject (简要说明，必填)

A short and imperative sentence describing the change.
用一句**祈使语气**的简短英文描述更改内容（不要用过去式或第三人称）。

✅ Use imperative mood (e.g. *add*, *fix*, *update*)
✅ 祈使语气，如 “add” 而非 “added” 或 “adds”
✅ Keep under **50 characters**
✅ 尽量控制在 **50 个字符以内**

Example / 示例：

```
feat(login): add password encryption
```

---

## 📄 4. Body (详细说明，可选)

Provide more detail about *what* and *why*, not *how*.
解释更改的内容和原因，而非实现细节。
Each line should wrap at 72 characters.
每行不要超过 72 个字符。

Example / 示例：

```
The previous login flow stored passwords in plain text.
This update introduces bcrypt for encryption to enhance security.
```

---

## 🔖 5. Footer (附注，可选)

Used for **breaking changes**, **issue references**, or additional notes.
用于标记**重大变更**、**关联 issue** 或 **补充说明**。

Example / 示例：

```
BREAKING CHANGE: The login API endpoint has been renamed.
Closes #128, #135
```

---

## 💡 Example Full Commit / 完整提交示例

```
feat(api): add JWT-based authentication

Introduce JWT tokens for secure user sessions.
Update middleware to verify tokens on each request.

Closes #45
```

---