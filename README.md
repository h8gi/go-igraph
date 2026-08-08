# go-igraph

igraph の C API を Go から利用するためのバインディングです。

## upstream API のカバレッジ

`make coverage` を実行すると、固定した upstream igraph リリースの header に名前が明記された
`IGRAPH_EXPORT` 関数と、
このリポジトリの production Go コードから実際に呼んでいる `C.igraph_*` 関数を比較し、
[`COVERAGE.md`](COVERAGE.md) を再生成します。対象バージョンと取得元は
[`coverage.json`](coverage.json) で明示しています。

型別 vector のように macro で生成される API は分母に含みません。また、この数値は
「Go API として完全に利用でき、十分にテストされている」ことを意味するものではなく、
実装候補を漏れなく把握するための機械的な一次指標です。upstream で廃止・改名された関数も
別一覧に出るため、バージョン更新時の移行作業にも利用できます。

ネットワークを使わずに確認する場合は、展開済みの igraph source tree を指定できます。

```sh
python3 tools/api_coverage.py --source /path/to/igraph
```

生成済みレポートとの差分確認（CI 向け）は `make coverage-check` で行えます。
