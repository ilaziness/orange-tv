# 如何编写 Git 提交信息

> 原文：[How to Write a Git Commit Message](https://chris.beams.io/git-commit) by Chris Beams

---

## 引言：为什么好的提交信息如此重要

如果你随意翻阅某个 Git 仓库的提交日志，会发现提交信息大多乱七八糟。以我早年向 Spring 项目提交代码时留下的这些"精彩之作"为例：

```
$ git log --oneline -5 --author cbeams --before "Fri Mar 26 2009"

e5f4b49 Re-adding ConfigurationPostProcessorTests after its brief removal in r814. @Ignore-ing the testCglibClassesAreLoadedJustInTimeForEnhancement() method as it turns out this was one of the culprits in the recent build breakage. The classloader hacking causes subtle downstream effects, breaking unrelated tests. The test method is still useful, but should only be run on a manual basis to ensure CGLIB is not prematurely classloaded, and should not be run as part of the automated build.
2db0f12 fixed two build-breaking issues: + reverted ClassMetadataReadingVisitor to revision 794 + eliminated ConfigurationPostProcessorTests until further investigation determines why it causes downstream tests to fail (such as the seemingly unrelated ClassPathXmlApplicationContextTests)
147709f Tweaks to package-info.java files
22b25e0 Consolidated Util and MutableAnnotationUtils classes into existing AsmUtils
7f96f57 polishing
```

再看看同一个仓库中[最近的提交](https://github.com/spring-projects/spring-framework/commits/5ba3db?author=philwebb)：

```
$ git log --oneline -5 --author pwebb --before "Sat Aug 30 2014"

5ba3db6 Fix failing CompositePropertySourceTests
84564a0 Rework @PropertySource early parsing logic
e142fd1 Add tests for ImportSelector meta-data
887815f Update docbook dependency and generate epub
ac8326d Polish mockito usage
```

你更愿意阅读哪一种？

前者长短不一、格式混乱；后者简洁一致。前者是不加约束的自然结果，后者则绝非偶然。

虽然很多仓库的日志看起来像前者，但也有例外。Linux 内核和 Git 本身就是很好的范例。再看看 Spring Boot，或者任何由 [Tim Pope](https://github.com/tpope/vim-pathogen/commits/master) 维护的仓库。

这些仓库的贡献者深知：**精心撰写的 Git 提交信息是向团队成员（以及未来的自己）传达变更上下文的最佳途径**。`diff` 能告诉你改了什么，但只有提交信息才能说清楚为什么这样改。Peter Hutterer [曾有精辟的论述](http://who-t.blogspot.co.at/2009/12/on-commit-messages.html)：

> 重建一段代码的上下文极其耗时。我们无法完全避免这件事，只能尽力减少它。提交信息正是做到这一点的绝佳手段，因此，一条提交信息是否写得好，可以直接体现出一名开发者是否是好的协作者。

如果你从未认真思考过什么是优秀的提交信息，很可能是因为你平时很少使用 `git log` 及相关工具。这里存在一个恶性循环：正因为提交历史混乱且不一致，人们才不愿花时间去查阅和维护它；而越是不去查阅维护，它就越发混乱不一致。

但一份精心维护的日志是美观且实用的宝藏。`git blame`、`revert`、`rebase`、`log`、`shortlog` 等子命令都会因此焕发生机。审查他人的提交和 Pull Request 变得有价值，且可以独立完成。理解几个月乃至几年前某件事发生的原因，不仅成为可能，而且变得高效。

一个项目的长期成功，在很大程度上依赖于其可维护性，而维护者手中最强大的工具之一，便是项目的提交日志。学会如何悉心打理它，是值得投入时间的事。起初可能觉得麻烦，但很快就会成为习惯，最终成为所有参与者的骄傲与生产力来源。

在本文中，我只讨论维护健康提交历史中最基础的一个要素：**如何撰写一条提交信息**。关于提交压缩（commit squashing）等其他重要实践，本文暂不涉及，或许会在后续文章中探讨。

大多数编程语言都有成熟的代码风格惯例，例如命名规范、格式规范等。当然，这些惯例也存在变体，但大多数开发者都认同：选定一种并坚持执行，远胜于人各一套所带来的混乱局面。

团队对待提交日志的态度也应如此。为了建立有用的版本历史，团队应首先就提交信息规范达成共识，至少明确以下三点：

- **风格**：格式标记语法、换行边界、语法、大小写、标点。把这些都明确规定下来，消除歧义，尽量保持简单。最终的成果将是一份格式高度一致、不仅赏心悦目，而且真正会被定期翻阅的日志。
- **内容**：提交信息正文（如果有的话）应该包含哪些信息？哪些又不该包含？
- **元数据**：如何引用 Issue 跟踪 ID、Pull Request 编号等信息？

好在关于什么是惯用的 Git 提交信息，已有成熟的行业约定可循，许多 Git 命令的运作方式也都隐含了这些约定。你无需另起炉灶，只需遵循下面的七条规则，便可以像专家一样提交代码。

---

## 优秀 Git 提交信息的七条规则

> 这些内容[已被](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html)[反复](https://www.git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project)[提及](https://github.com/torvalds/subsurface-for-dirk/blob/master/README.md)——这里只是做一个系统性的梳理。

1. **用空行将标题与正文分隔**
2. **标题行限制在 50 个字符以内**
3. **标题行首字母大写**
4. **标题行末尾不加句号**
5. **标题行使用祈使语气**
6. **正文每行在 72 个字符处换行**
7. **用正文解释 What 和 Why，而非 How**

例如：

```
用约 50 个字符概括变更内容

如有必要，在此处撰写更详细的说明文字，每行约 72 个字符换行。
在某些场景下，第一行会被视为提交的主题（subject），其余文字
视为正文（body）。将摘要与正文隔开的空行至关重要（除非你完全
省略正文）；如果两者连在一起，log、shortlog、rebase 等工具
可能会产生混淆。

说明本次提交所解决的问题。重点阐述为何做出这一改动，而非如何
实现（代码本身已能说明这一点）。这一改动是否有副作用或其他不
直观的后果？这里是解释的好地方。

后续段落之间以空行分隔。

- 使用要点列表也可以
- 通常使用连字符或星号作为列表符号，前面加一个空格，各条目之
  间可加空行，具体格式因惯例而异

如果使用 Issue 跟踪系统，请在最后引用相关条目，例如：

Resolves: #123
See also: #456, #789
```

---

### 1. 用空行将标题与正文分隔

来自 `git commit` 的 [man 手册页](https://www.kernel.org/pub/software/scm/git/docs/git-commit.html)：

> 虽然不是必须的，但最好以一行简短（不超过 50 个字符）的摘要开头，后跟一个空行，再是更详细的描述。提交信息中第一个空行之前的文字，会被视为提交的标题（commit title），该标题会在整个 Git 中被广泛使用。

首先，并非每次提交都需要标题和正文两部分。有时一行就够了，尤其是当改动非常简单、无需额外上下文时。例如：

```
Fix typo in introduction to user guide
```

（修正用户手册引言中的错别字）

无需再多说什么；如果读者想知道是什么错别字，直接查看改动本身即可，比如用 `git show`、`git diff` 或 `git log -p`。

如果你在命令行中提交这样简单的内容，使用 `git commit` 的 `-m` 选项很方便：

```
$ git commit -m "Fix typo in introduction to user guide"
```

但是，当一次提交需要一定的解释和上下文时，就需要撰写正文了。例如：

```
Derezz the master control program

MCP turned out to be evil and had become intent on world domination.
This commit throws Tron's disc into MCP (causing its deresolution)
and turns it back into a chess game.
```

带有正文的提交信息不适合用 `-m` 选项来写。此时最好在合适的文本编辑器中撰写。如果你还没有为 Git 命令行配置好编辑器，请参阅 [Pro Git 的这一章节](https://git-scm.com/book/en/v2/Customizing-Git-Git-Configuration)。

无论如何，在浏览日志时，标题与正文的分隔都会带来明显的好处。完整的日志条目如下：

```
$ git log
commit 42e769bdf4894310333942ffc5a15151222a87be
Author: Kevin Flynn <kevin@flynnsarcade.com>
Date:   Fri Jan 01 00:00:00 1982 -0200

 Derezz the master control program

 MCP turned out to be evil and had become intent on world domination.
 This commit throws Tron's disc into MCP (causing its deresolution)
 and turns it back into a chess game.
```

执行 `git log --oneline` 时，只输出标题行：

```
$ git log --oneline
42e769 Derezz the master control program
```

执行 `git shortlog` 时，按用户分组，同样只显示标题行：

```
$ git shortlog
Kevin Flynn (1):
      Derezz the master control program

Alan Bradley (1):
      Introduce security program "Tron"

Ed Dillinger (3):
      Rename chess program to "MCP"
      Modify chess program
      Upgrade chess program

Walter Gibbs (1):
      Introduce protoype chess program
```

Git 中还有许多其他场景会用到标题行与正文的区分——但如果没有中间那个空行，所有这些功能都无法正常运作。

---

### 2. 标题行限制在 50 个字符以内

50 个字符并非硬性限制，只是一条经验法则。将标题行长度控制在这个范围内，既保证了可读性，也迫使作者花心思找到描述变更的最简洁方式。

> **译注**：如果你总是难以概括提交的内容，可能说明这次提交包含了太多变更。尽量做到[原子化提交（atomic commits）](https://www.freshconsulting.com/atomic-commits/)。

GitHub 的 UI 完全遵循这些约定。超过 50 个字符时会发出警告；超过 72 个字符的标题行会被截断并显示省略号。

因此，目标是 50 个字符，72 个字符是硬性上限。

---

### 3. 标题行首字母大写

顾名思义。**所有标题行的首字母都应大写。**

例如，应写：

- `Accelerate to 88 miles per hour`

而非：

- `accelerate to 88 miles per hour`

---

### 4. 标题行末尾不加句号

标题行末尾的标点是多余的。况且，当你尽力将其控制在 [50 个字符以内](https://cbea.ms/posts/git-commit/)时，每一个字符都弥足珍贵。

例如，应写：

- `Open the pod bay doors`

而非：

- `Open the pod bay doors.`

---

### 5. 标题行使用祈使语气

**祈使语气**，即像发出命令或指示一样来书写。几个日常例子：

- `Clean your room`（打扫你的房间）
- `Close the door`（关上门）
- `Take out the trash`（倒垃圾）

你现在正在阅读的这七条规则，每一条都是用祈使语气写成的（"Wrap the body at 72 characters" 等等）。

祈使语气有时听起来略显生硬，这也是日常交流中我们不常使用它的原因。但它非常适合 Git 提交信息的标题行。原因之一是：**Git 本身在代你创建提交时，使用的就是祈使语气**。

例如，`git merge` 创建的默认提交信息是：

```
Merge branch 'myfeature'
```

`git revert` 时：

```
Revert "Add the thing with the stuff"

This reverts commit cc87791524aedd593cff5a74532befe7ab69ce9d.
```

在 GitHub 上点击 "Merge" 按钮合并 Pull Request 时：

```
Merge pull request #123 from someuser/somebranch
```

因此，用祈使语气撰写提交信息，实际上是在遵循 Git 自身的内置惯例。例如：

- `Refactor subsystem X for readability`（为了可读性重构子系统 X）
- `Update getting started documentation`（更新入门文档）
- `Remove deprecated methods`（移除已废弃的方法）
- `Release version 1.0.0`（发布 1.0.0 版本）

这种写法一开始可能有些别扭，因为我们更习惯使用陈述语气来描述事实。这也是提交信息经常被写成这样的原因：

- `Fixed bug with Y`
- `Changing behavior of X`

有时提交信息甚至被写成对内容的描述：

- `More fixes for broken stuff`
- `Sweet new API methods`

为了消除歧义，这里有一个简单的规则，可以帮你每次都写对：

**一条格式正确的 Git 提交标题行，应该能够完成以下句子：**

> **If applied, this commit will** \_\_\_\_\_your subject line here\_\_\_\_\_
>
> （如果应用此提交，它将……）

例如：

- If applied, this commit will **refactor subsystem X for readability**
- If applied, this commit will **update getting started documentation**
- If applied, this commit will **remove deprecated methods**
- If applied, this commit will **release version 1.0.0**
- If applied, this commit will **merge pull request #123 from user/branch**

注意，这个公式对非祈使语气的写法并不适用：

- If applied, this commit will ~~fixed bug with Y~~
- If applied, this commit will ~~changing behavior of X~~
- If applied, this commit will ~~more fixes for broken stuff~~
- If applied, this commit will ~~sweet new API methods~~

> **译注**：祈使语气的要求**仅适用于标题行**。正文部分可以放宽要求。

---

### 6. 正文每行在 72 个字符处换行

Git 不会自动折行。在撰写提交信息正文时，你需要注意右边界，**手动换行**。

推荐在 72 个字符处换行，这样 Git 在缩进文字时仍有足够的空间，同时确保总体不超过 80 个字符。

好的文本编辑器可以提供帮助。例如，可以轻松配置 Vim 在撰写 Git 提交信息时自动在 72 个字符处折行。然而传统上，IDE 对提交信息的智能折行支持一直很差（尽管近年来 IntelliJ IDEA 已经有所[改善](https://youtrack.jetbrains.com/issue/IDEA-53615)）。

---

### 7. 用正文解释 What 和 Why，而非 How

来自 Bitcoin Core 的[这条提交](https://github.com/bitcoin/bitcoin/commit/eb0b56b19017ab5c16c745e6da39c53126924ed6)是解释"改了什么"和"为什么改"的绝佳范例：

```
commit eb0b56b19017ab5c16c745e6da39c53126924ed6
Author: Pieter Wuille <pieter.wuille@gmail.com>
Date:   Fri Aug 1 22:57:55 2014 +0200

   Simplify serialize.h's exception handling

   Remove the 'state' and 'exceptmask' from serialize.h's stream
   implementations, as well as related methods.

   As exceptmask always included 'failbit', and setstate was always
   called with bits = failbit, all it did was immediately raise an
   exception. Get rid of those variables, and replace the setstate
   with direct exception throwing (which also removes some dead code).

   As a result, good() is never reached after a failure (there are
   only 2 calls, one of which is in tests), and can just be replaced
   by !eof().

   fail(), clear(n) and exceptions() are just never called. Delete
   them.
```

查看[完整 diff](https://github.com/bitcoin/bitcoin/commit/eb0b56b19017ab5c16c745e6da39c53126924ed6)，想想作者在此时此刻花时间提供这些上下文，为未来的维护者节省了多少时间。如果他没有写下来，这些信息很可能就永远消失了。

在大多数情况下，你可以省略关于"如何实现"的细节。代码在这方面通常是自解释的（如果代码太复杂以至于需要文字解释，那是源码注释该做的事）。**专注于阐明你做出这一改动的原因**——改动之前的状况（以及为何不妥），改动之后的效果，以及你为何选择这种方式来解决问题。

未来感谢你的维护者，可能就是你自己！

---

## 延伸建议

### 学会热爱命令行，放下 IDE

基于 Git 子命令数量之多，拥抱命令行是明智之举。Git 的能力极其强大；IDE 亦然，但两者各有所长。我每天使用 IDE（IntelliJ IDEA），也大量使用过其他 IDE（Eclipse），但我从未见过哪款 IDE 的 Git 集成能在易用性和功能性上媲美命令行（前提是你已经熟悉了它）。

某些 Git 相关的 IDE 功能确实无价，比如删除文件时自动调用 `git rm`，重命名文件时正确处理 `git` 操作。但当你开始尝试通过 IDE 来执行提交、合并、变基（rebase），或进行复杂的历史分析时，一切便开始崩塌。

若要发挥 Git 的全部威力，命令行才是正途。

无论你使用 Bash、Zsh 还是 PowerShell，都有 [Tab 补全脚本](https://git-scm.com/book/en/v2/Appendix-A%3A-Git-in-Other-Environments-Git-in-PowerShell)可以大幅降低记忆子命令和选项的负担。

### 阅读《Pro Git》

[《Pro Git》](https://git-scm.com/book/en/v2)一书可在线免费阅读，内容极佳，好好利用它吧！

---

_题图来源：[xkcd](https://xkcd.com/1296/)_
