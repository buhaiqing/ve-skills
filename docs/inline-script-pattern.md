# 内联脚本模式（Inline Script Pattern）

> **用途**：将 Python 函数的函数体序列化为 `python3 -c '...'` 内联脚本，避免创建临时 .py 文件。
> **使用场景**：`scripts/validate_local.py` 中的 `_inline_script()` 辅助函数。
> 沉淀自 2026-07-05 validate_local.py 开发过程中的两次修复经历。

---

## 问题背景

`validate_local.py` 需要运行多个检查步骤。其中一些步骤逻辑简单（如遍历目录检查文件），但需要作为独立子进程执行（与主进程隔离副作用）。方案选择：

| 方案 | 问题 |
|------|------|
| 写成独立 `.py` 文件 | 文件膨胀，每个小检查一个文件 |
| 通过 `python3 -c '...'` 内联 | 字符串转义、调试困难 |
| **`_inline_script()` 方案** | **从当前文件提取函数体，自动序列化为 `-c` 脚本** |

## 实现

参考 `scripts/validate_local.py` 中的 `_inline_script()`：

```python
def _inline_script(fn):
    import inspect, textwrap
    source = inspect.getsource(fn)
    lines = source.splitlines()
    body = textwrap.indent(textwrap.dedent("\n".join(lines[1:])), "    ")
    return (
        "import sys\n"
        "from pathlib import Path\n"
        "def main():\n"
        "    root = Path.cwd()\n"
        f"{body}\n"
        'sys.exit(main())\n'
    )
```

输入一个带 `root: Path` 签名的函数，输出一个可执行的 `python3 -c` 脚本字符串。

---

## 已知约束（踩过的坑）

### 1. `python3 -c` 中没有 `__file__`

**问题**：初始实现用 `Path(__file__).resolve().parents[1]` 获取项目根目录，但在 `-c` 模式下 `__file__` 未定义。

**解法**：`run_step()` 已设置 `cwd=root`，所以内联脚本直接用 `Path.cwd()` 获取当前工作目录。

### 2. 模块级 `return` 非法

**问题**：函数体内的 `return 1` 和 `return 0` 在模块级代码中不合法。

**解法**：将整个内联脚本包装在 `main()` 函数中，用 `sys.exit(main())` 获取退出码。

```python
# 生成的脚本结构
def main():
    root = Path.cwd()
    ...  # 原函数体
sys.exit(main())
```

### 3. 函数签名参数丢失

**问题**：被序列化的函数签名如 `def _check(root: Path)` 的 `root` 参数被丢弃（因为跳过了 `def` 行）。

**解法**：包装器在 `main()` 顶部注入 `root = Path.cwd()`，函数体直接引用 `root`。

### 4. 缩进边界

**问题**：函数体在源文件中缩进 4 层，但在 `main()` 内部需缩进 8 层（`main()` 自身 4 层 + 函数体 4 层）。

**解法**：先 `textwrap.dedent()` 去缩进，再用 `textwrap.indent(..., "    ")` 加 4 层。

### 5. import 位置

**问题**：函数体内可能使用 `import re` 等模块级导入，在 `main()` 内部也合法。

**解法**：无特殊处理，Python 函数内 import 正常工作。

---

## 适用范围

| 适合 | 不适合 |
|------|--------|
| 单文件、无外部依赖的检查逻辑 | 需要安装的第三方包 |
| 遍历目录、检查文件内容的简单操作 | 需要长时间运行的守护进程 |
| 作为子进程隔离运行（cwd 已知） | 需要交互式输入的逻辑 |
| 与主进程共享同一个 repo 的工具脚本 | 需要访问 `__file__` 所在目录 |

## 演进的替代方案

如果未来脚本数量增长，可以考虑：

1. **多文件拆分**：每个检查独立为 `scripts/check_*.py`，`validate_local.py` 只编排不内含逻辑
2. **`subprocess.run(sys.executable, "-c", ...)`** 保持当前模式，但将 `_inline_script` 改为从独立文件中读取代码而非 `inspect.getsource`
3. **pytest 替代**：如果检查逻辑需要 mock/parametrize
