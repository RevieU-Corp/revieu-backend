# 交接文档 — QB 商家/菜品/优惠券管理(#223-227)

**暂停时间点:** 后端(Task 1-9)全部完成;前端 Task 10-11 已完成代码;Task 11 评审发现 2 个 Important 级别的样式问题**尚未修复**,是暂停时最后的未完成状态。Task 12-18 尚未开始。

## 快速定位

| 内容 | 位置 |
|---|---|
| Spec(设计文档) | `revieu-core-api-go/docs/superpowers/specs/2026-08-21-qb-merchant-dish-coupon-design.md` |
| Plan(实施计划,18 个任务的完整代码) | `revieu-core-api-go/docs/superpowers/plans/2026-08-21-qb-merchant-dish-coupon.md` |
| 执行台账(每个任务的完整记录、所有裁决) | `revieu-core-api-go/.worktrees/qb-merchant-dish-coupon/.superpowers/sdd/2026-08-21-qb-merchant-dish-coupon/progress.md` |
| 后端 worktree | `/home/paul2/workspace/repos/revieu-core-api-go/.worktrees/qb-merchant-dish-coupon`,分支 `feature/qb-merchant-dish-coupon` |
| 前端 worktree | `/home/paul2/workspace/repos/revieu-web/.worktrees/qb-merchant-dish-coupon`,分支 `feature/qb-merchant-dish-coupon` |

两个分支都已经 **push 到各自的 origin**(见文末)。

## 整体进度

- ✅ **后端 Task 1-9 全部完成**,`go build ./... && go vet ./... && go test ./...` 全绿,已独立复核。最终 commit:`26c0905`。
- ✅ **前端 Task 10 完成**(`storeProfileService` 新增 `createStore`/`activateStore`/`deactivateStore`)。
- ⚠️ **前端 Task 11 代码已完成但未收尾**(StoreProfile.tsx 新建商家表单 —— 这是 #224 的核心 bug 修复)。评审判定 "Approved" 但列了 2 条 Important(不影响功能,纯样式):
  1. 新的"创建门店"按钮用了 `bg-blue-600`,和全文件其他主按钮的品牌黄色(`#FFBC0D` / `bg-yellow-500`)不一致
  2. 创建表单里的错误提示用红色底(`bg-red-50 text-red-600`),其他地方用琥珀色底(`bg-amber-50 text-amber-800`)
  已经裁决"现在就修"(demo 前的第一屏,改动成本很低),但**修复还没有派发**——这是暂停时的确切断点。
- ❌ **前端 Task 12-18 尚未开始**:门店启用/禁用、菜品管理、优惠券创建表单、优惠券横向列表、把 Dashboard 里假的 CouponManager 换成真实调用。

## 恢复执行的步骤

1. 用 `superpowers:subagent-driven-development` 技能恢复,先读一遍 `progress.md`(台账),确认从 "Task 11 待修复" 这里继续,不要重新跑 Task 1-10。
2. **Task 11 的修复**:两处样式一行 CSS class 替换,不需要重新设计,直接改:
   - `bg-blue-600 hover:bg-blue-700` → 参照文件里其他主按钮的样式(`bg-yellow-500 hover:bg-yellow-600` + `style={{ backgroundColor: '#FFBC0D' }}`)
   - 创建表单的状态提示框从红色改成琥珀色(`border-amber-200 bg-amber-50 text-amber-800`,和文件里第 614-618 行一致)
   - 修完跑一次 scoped re-review,通过后记 `Task 11: complete`。
3. 之后按 Plan 文档顺序继续 Task 12 → 18,每个任务:派实现子代理 → 我独立复核(build/test 或 tsc)→ 派评审子代理 → 有 Important/Critical 就进入修复循环 → 记账本 → 下一个。
4. 全部 18 个任务完成后:按技能流程做一次全分支的最终评审(`superpowers:requesting-code-review` 的 code-reviewer,用最强模型),修完最后一轮发现的问题,然后用 `superpowers:finishing-a-development-branch` 决定怎么合并这两个 worktree 分支。

## 过程中做出的裁决(Rulings)——需要你知道的

这些是我代替你做的决定,按发生顺序列出,附带"如果我判断错了,代价是什么":

1. **Task 9 spec 缺口(已修复,提交前)**:优惠券状态计算函数(`computeStatus`)在原计划里写了但从没被任何接口真正调用。改成导出为 `ComputeStatus` 并在 Task 9 的所有优惠券响应接口里包一层 wrapper 使用。代价:一个纯新增的小类型,容易撤销。

2. **Task 4 路由接入(已修复,提交前)**:原计划让实现者"猜"路由注册的具体位置,我直接查了 `router.go` 给出确切的变量名(`api`)和 import 顺序。代价:无,只是更精确。

3. **Task 14 前端接入方式(已修复,提交前)**:原计划写的是直接 import 页面文件,但这个代码库实际用的是 barrel 文件转出口模式。已改成走 `dishes/index.ts` → `features/merchant/index.ts` → `App.tsx` 这条正确路径。代价:多两个一行文件,容易改回去。

4. **Task 1 实现子代理跑偏(已修复,你已批准)**:第一次派发时,实现子代理把改动提交到了主仓库的 `dev` 分支,而不是隔离的 worktree(这个 harness 里 `cd` 不会跨 bash 调用保持)。我用 `git reset --soft` + 恢复两个受影响文件的方式把 `dev` 分支干净地复原,验证过完全没影响你在 `dev` 上原有的未提交改动,worktree 本身也确认没被污染。之后重新派发时在指令里加了强制校验。代价:无——那次误提交从没 push 过,git 对象还在,理论上可以用 `git fsck` 找回(commit SHA `434db2f`)。

5. **Task 8 评审发现(已修复)**:编辑优惠券(`UpdateForStore`)没有校验开始时间不能晚于结束时间,创建时有这个校验但编辑时漏了。这是计划文档本身的疏漏,不是实现者的偏差。判断为"现在就修"因为这关系到 #227 编辑优惠券这个功能是否可靠。代价:几行校验逻辑,容易撤销。

6. **Task 9 评审发现(已修复)**:编辑优惠券接口对 `status` 字段没有白名单校验,理论上能被直接传参改成"售罄"/"过期"/"未开始"这种本该只由系统计算的值,破坏了"派生状态不落库"这个设计前提。同样是计划文档的疏漏。判断为"现在就修",因为这个漏洞正好是 Task 7+9 整套状态计算设计要保护的东西。代价:一个校验分支,容易撤销。

7. **Task 11 评审发现(已裁决,尚未执行修复)**:见上文"整体进度"部分,已经判断为"现在就修"但还没派发子代理去做。

## 记录在案但故意不修的次要问题(Minor,已延后)

这些都记在 `progress.md` 里,不影响功能,最终整体评审时会再统一看一遍要不要处理:

- Task 1:`updateStatusOwned`(发布触发自动认证)路径没有专门测试;`verifyMerchantIfEnabled` 在事务提交后才跑,理论上有极窄的"店已建但报错"边界(计划本身这么设计的)。
- Task 3/6:测试用的 sqlite 在查不到记录时会打印 GORM 自带的 "record not found" 错误日志——这是 `coupon` 包早就有的既有模式,不是这次引入的。
- Task 6:菜品图片查询失败时(不只是查不到,任何 DB 错误)会被静默吞掉——计划本身这么写的,风险很窄。
- Task 7:`ListForMerchant` 和 `DeleteForStore` 有一段一模一样的"校验归属权"代码,是计划明确要求先不重构,以后可以抽个公共函数。
- Task 8:编辑优惠券对 `TotalQuantity` 允许 0(创建时要求 >0)——计划原文如此;缺一个"优惠券不属于这个门店"的编辑场景测试;`ValidFrom`/`ValidUntil` 显式清空为 null 的双重指针分支没测试(代码走查确认逻辑对)。
- Task 9:新加的 `model` import 没有按字母序排(base 文件本来也没排好,不是新引入的问题);通过 HTTP API 目前没法显式把 `valid_from`/`valid_until` 清空成 null(service 层支持,DTO 层没打通,没有调用方需要这个能力);`DeleteStoreCoupon` 还在用旧的内联 `ParseInt`,没有换成新的 `parseStoreAndCouponID` 辅助函数(为了保持 diff 最小,评审认可这个取舍);`apps/core/docs/`(swagger 文档)没有为新接口重新生成,是跨任务的遗留事项。
- Task 11:表单输入框缺少和其他地方一致的 focus 高亮样式;label 的下边距 `mb-1` 和其他地方的 `mb-2` 不一致。

## 环境相关的重要提醒

- **`AUTO_VERIFY_NEW_MERCHANTS`** 环境变量控制新商家是否自动标记为已认证(demo 专用)。默认关闭,需要在 `revieu-dev` 环境手动打开才能让新建的商家立刻在客户端可见。**这个开关目前还没有在任何环境里配置**,记得部署前处理。
- 两个 worktree 都要记得跑一次部署 / 重启后端服务,新代码才会在 `revieu-dev` 生效——这个仓库有 CI/CD(推到 `dev` 分支触发),但 `feature/qb-merchant-dish-coupon` 分支现在还没有合并回 `dev`,合并方式留到最后 `finishing-a-development-branch` 阶段再定。
