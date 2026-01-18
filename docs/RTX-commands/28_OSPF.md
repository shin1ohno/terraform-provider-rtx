# 第28章: OSPF

> 元PDFページ: 452-466

---

第 28 章
OSPF
OSPF はインテリアゲートウェイプロトコルの一種で、 グラフ理論をベースとしたリンク状態型の動的ルーティング
プロトコルである。
28.1 OSPF の有効設定
[書式 ]
ospf configure  refresh
[説明 ]
OSPF 関係の設定を有効にする。 OSPF 関係の設定を変更したら、ルーターを再起動するか、あるいはこのコマンド
を実行しなくてはいけない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.2 OSPF の使用設定
[書式 ]
ospf use use
no ospf use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on OSPF を使用する
off OSPF を使用しない
• [初期値 ] : off
[説明 ]
OSPF を使用するか否かを設定する。
[ノート ]
以下の機能はまだサポートされていない。
• NSSA (RFC1587)
• OSPF over demand circuit (RFC1793)
• OSPF MIB
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.3 OSPF による経路の優先度設定
[書式 ]
ospf preference  preference
no ospf preference  [preference ]
[設定値及び初期値 ]
•preference
• [設定値 ] : OSPF による経路の優先度  (1 以上の数値  )
• [初期値 ] : 2000452 | コマンドリファレンス  | OSPF

[説明 ]
OSPF による経路の優先度を設定する。優先度は  1 以上の数値で表され、数字が大きい程優先度が高い。 OSPF と
RIP など複数のプロトコルで得られた経路が食い違う場合には、 優先度が高い方が採用される。優先度が同じ場合に
は時間的に先に採用された経路が有効となる。
[ノート ]
静的経路の優先度は  10000 で固定である。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.4 OSPF のルーター ID 設定
[書式 ]
ospf router  id router-id
no ospf router  id [router-id ]
[設定値及び初期値 ]
•router_id
• [設定値 ] : IP アドレス
• [初期値 ] : -
[説明 ]
OSPF のルーター ID を指定する。
[ノート ]
ルーター  ID が本コマンドで設定されていないときは、以下のインタフェースに付与されているプライマリ  IPv4 ア
ドレスのいずれかが自動的に選択され、ルーター  ID として使用させれる。
• LAN インタフェース
• LOOPBACK インタフェース
• PP インタフェース
なお、プライマリ  IPv4 アドレスが付与されたインタフェースがない場合は初期値は設定されない。
意図しない  IP アドレスがルーター  ID として使用されることを防ぐため、本コマンドにより明示的にルーター  ID を
指定することが望ましい。
OSPF と BGP-4 とを併用する場合、本コマンドか  bgp router id コマンドのいずれか一方を設定する。
以下のリビジョンでは、本コマンドと  bgp router id コマンドの両方を設定することができるが、必ず同一のルーター
ID を指定する必要がある。
機種 リビジョン
vRX シリーズ すべてのリビジョン
RTX1300 すべてのリビジョン
RTX1220 すべてのリビジョン
RTX1210 Rev.14.01.16
RTX5000 Rev.14.00.21
RTX3510 すべてのリビジョン
RTX3500 Rev.14.00.21
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.5 OSPF で受け取った経路をルーティングテーブルに反映させるか否かの設定
[書式 ]
ospf export  from  ospf [filter filter_num ...]
no ospf export  from  ospf [filter filter_num ...]コマンドリファレンス  | OSPF | 453

[設定値及び初期値 ]
•filter_num
• [設定値 ] : ospf export filter  コマンドのフィルタ番号
• [初期値 ] : すべての経路がルーティングテーブルに反映される
[説明 ]
OSPF で受け取った経路をルーティングテーブルに反映させるかどうかを設定する。指定したフィルタに一致する
経路だけがルーティングテーブルに反映される。コマンドが設定されていない場合または  filter キーワード以降を
省略した場合には、すべての経路がルーティングテーブルに反映される。
[ノート ]
フィルタ番号は 101個まで設定できる。
このコマンドは  OSPF のリンク状態データベースには影響を与えない。つまり、 OSPF で他のルーターと情報をやり
取りする動作としては、このコマンドがどのように設定されていても変化は無い。 OSPF で計算した経路が、実際に
パケットをルーティングするために使われるかどうかだけが変わる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.6 外部プロトコルによる経路導入
[書式 ]
ospf import  from  protocol  [filter filter_num ...]
no ospf import  from  protocol  [filter filter_num ...]
[設定値及び初期値 ]
•protocol  : OSPF の経路テーブルに導入する外部プロトコル
• [設定値 ] :
設定値 説明
static 静的経路  ( implicit 経路は含まない  )
rip RIP
bgp BGP
• [初期値 ] : -
•filter_num
• [設定値 ] : フィルタ番号
• [初期値 ] : -
[説明 ]
OSPF の経路テーブルに外部プロトコルによる経路を導入するかどうかを設定する。導入された経路は外部経路と
して他の  OSPF ルーターに広告される。
filter_num  はospf import filter  コマンドで定義したフィルタ番号を指定する。外部プロトコルから導入されようとす
る経路は指定したフィルタにより検査され、フィルタに該当すればその経路は  OSPF に導入される。該当するフィ
ルタがない経路は導入されない。また、 filter キーワード以降を省略した場合には、すべての経路が  OSPF に導入さ
れる
経路を広告する場合のパラメータであるメトリック値、メトリックタイプ、タグは、フィルタの検査で該当した ospf
import filter  コマンドで指定されたものを使う。 filter キーワード以降を省略した場合には、 以下のパラメータを使用
する。
• metric=1
• type=2
• tag=1
[ノート ]
フィルタ番号は、 vRX シリーズ、 RTX5000 、RTX3510 、RTX3500 は 300 個、他の機種は  101 個まで設定できる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830454 | コマンドリファレンス  | OSPF

28.7 OSPF で受け取った経路をどう扱うかのフィルタの設定
[書式 ]
ospf export  filter  filter_num  [nr] kind ip_address /mask ...
no ospf export  filter  filter_num  [...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号
• [初期値 ] : -
•nr : フィルタの解釈の方法
• [設定値 ] :
設定値 説明
not フィルタに該当しない経路を導入する
reject フィルタに該当した経路を導入しない
省略時 フィルタに該当した経路を導入する
• [初期値 ] : -
•kind : フィルタ種別
• [設定値 ] :
設定値 説明
include 指定したネットワークアドレスに含まれる経路  ( ネ
ットワークアドレス自身を含む  )
refines 指定したネットワークアドレスに含まれる経路  ( ネ
ットワークアドレス自身を含まない  )
equal 指定したネットワークアドレスに一致する経路
• [初期値 ] : -
•ip_address/mask
• [設定値 ] : ネットワークアドレスをあらわす  IP アドレスとマスク長
• [初期値 ] : -
[説明 ]
OSPF により他の  OSPF ルーターから受け取った経路を経路テーブルに導入する際に適用するフィルタを定義する。
このコマンドで定義したフィルタは、 ospf export from ospf  コマンドの  filter 項で指定されてはじめて効果を持つ。
ip_address/mask  では、ネットワークアドレスを設定する。これは、複数設定でき、経路の検査時にはそれぞれのネ
ットワークアドレスに対して検査を行う。
nr が省略されている場合には、一つでも該当するフィルタがある場合には経路が導入される。
not 指定時には、すべての検査でフィルタに該当しなかった場合に経路が導入される。 reject 指定時には、一つでも
該当するフィルタがある場合には経路が導入されない
kind では、経路の検査方法を設定する。
includeネットワークアドレスと一致する経路および、ネットワ
ークアドレスに含まれる経路が該当となる
refinesネットワークアドレスに含まれる経路が該当となるが、
ネットワークアドレスと一致する経路が含まれない
equalネットワークアドレスに一致する経路だけが該当とな
る
[ノート ]
not 指定のフィルタを ospf export from ospf  コマンドで複数設定する場合には注意が必要である。 not 指定のフィル
タに合致するネットワークアドレスは、そのフィルタでは導入するかどうかが決定しないため、次のフィルタで検コマンドリファレンス  | OSPF | 455

査されることになる。そのため、例えば、以下のような設定ではすべての経路が導入されることになり、フィルタ
の意味が無い。
ospf export from ospf filter 1 2
ospf export filter 1 not equal 192.168.1.0/24
ospf export filter 2 not equal 192.168.2.0/24
1 番のフィルタでは、 192.168.1.0/24 以外の経路を導入し、 2 番のフィルタで  192.168.2.0/24 以外の経路を導入してい
る。つまり、経路  192.168.1.0/24 は 2 番のフィルタにより、経路  192.168.2.0/24 は 1 番のフィルタにより導入される
ため、導入されない経路は存在しない。
経路  192.168.1.0/24 と経路  192.168.2.0/24 を導入したくない場合には以下のような設定を行う必要がある。
ospf export from ospf filter 1
ospf export filter 1 not equal 192.168.1.0/24 192.168.2.0/24
あるいは
ospf export from ospf filter 1 2 3
ospf export filter 1 reject equal 192.168.1.0/24
ospf export filter 2 reject equal 192.168.2.0/24
ospf export filter 3 include 0.0.0.0/0
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.8 外部経路導入に適用するフィルタ定義
[書式 ]
ospf import  filter  filter_num  [nr] kind ip_address /mask ... [parameter ...].
no ospf import  filter  filter_num  [[not] kind ip_address /mask ... [parameter ...]]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号
• [初期値 ] : -
•nr : フィルタの解釈の方法
• [設定値 ] :
設定値 説明
not フィルタに該当しない経路を広告する
reject フィルタに該当した経路を広告しない
省略時 フィルタに該当した経路を広告する
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
include 指定したネットワークアドレスに含まれる経路  ( ネ
ットワークアドレス自身を含む  )
refines 指定したネットワークアドレスに含まれる経路  ( ネ
ットワークアドレス自身は含まない  )
equal 指定したネットワークアドレスに一致する経路
• [初期値 ] : -
•ip_address/mask
• [設定値 ] : ネットワークアドレスをあらわす  IP アドレスとマスク長
• [初期値 ] : -
•parameter  : 外部経路を広告する場合のパラメータ
• [設定値 ] :456 | コマンドリファレンス  | OSPF

設定値 説明
metric メトリック値  (0..16777215)
type メトリックタイプ  (1..2)
tag タグの値  (0..4294967295)
• [初期値 ] : -
[説明 ]
OSPF の経路テーブルに外部経路を導入する際に適用するフィルタを定義する。このコマンドで定義したフィルタ
は、 ospf import from  コマンドの  filter 項で指定されてはじめて効果を持つ。
ip_address/mask  では、ネットワークアドレスを設定する。これは、複数設定でき、経路の検査時にはそれぞれのネ
ットワークアドレスに対して検査を行い、 1 つでも該当するものがあればそれが適用される。
nr が省略されている場合には、一つでも該当するフィルタがある場合には経路を広告する。 not 指定時には、すべて
の検査でフィルタに該当しなかった場合に経路を広告する。 reject 指定時には、一つでも該当するフィルタがある場
合には経路を広告しない。
kind では、経路の検査方法を設定する。
includeネットワークアドレスと一致する経路および、ネットワ
ークアドレスに含まれる経路が該当となる
refinesネットワークアドレスに含まれる経路が該当となるが、
ネットワークアドレスと一致する経路が含まれない
equalネットワークアドレスに一致する経路だけが該当とな
る
kind の前に  not キーワードを置くと、該当 /非該当の判断が反転する。例えば、 not equal では、ネットワークアドレ
スに一致しない経路が該当となる
parameter  では、該当した経路を  OSPF の外部経路として広告する場合のパラメータとして、メトリック値、メトリ
ックタイプ、タグがそれぞれ  metric、type、tag により指定できる。これらを省略した場合には、以下の値が採用さ
れる。
• metric=1
• type=2
• tag=1
[ノート ]
not 指定のフィルタを ospf import from  コマンドで複数設定する場合には注意が必要である。 not 指定のフィルタに
合致するネットワークアドレスは、そのフィルタでは導入するかどうかが決定しないため、次のフィルタで検査さ
れることになる。そのため、例えば、以下のような設定ではすべての経路が広告されることになり、フィルタの意
味が無い。
ospf import from static filter 1 2
ospf import filter 1 not equal 192.168.1.0/24
ospf import filter 2 not equal 192.168.2.0/24
1 番のフィルタでは、 192.168.1.0/24 以外の経路を広告し、 2 番のフィルタで  192.168.2.0/24 以外の経路を広告してい
る。つまり、経路  192.168.1.0/24 は 2 番のフィルタにより、経路  192.168.2.0/24 は 1 番のフィルタにより広告される
ため、広告されない経路は存在しない。
経路  192.168.1.0/24 と経路  192.168.2.0/24 を広告したくない場合には以下のような設定を行う必要がある。
ospf import from static filter 1
ospf import filter 1 not equal 192.168.1.0/24 192.168.2.0/24
あるいは
ospf import from static filter 1 2 3
ospf import filter 1 reject equal 192.168.1.0/24
ospf import filter 2 reject equal 192.168.2.0/24
ospf import filter 3 include 0.0.0.0/0コマンドリファレンス  | OSPF | 457

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.9 OSPF エリア設定
[書式 ]
ospf area area [auth= auth] [stub [cost= cost]]
no ospf area area [auth= auth] [stub [cost= cost]]
[設定値及び初期値 ]
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値 非バックボーンエリア
IP アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : -
•auth
• [設定値 ] :
設定値 説明
text プレーンテキスト認証
md5 MD5 認証
• [初期値 ] : 認証は行わない
•stub : スタブエリアであることを指定する。
• [初期値 ] : スタブエリアではない
•cost
• [設定値 ] : 1 以上の数値
• [初期値 ] : -
[説明 ]
OSPF エリアを設定する。
cost は 1 以上の数値で、 エリア境界ルーターがエリア内に広告するデフォルト経路のコストとして使われる。 cost を
指定しないとデフォルト経路の広告は行われない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.10 エリアへの経路広告
[書式 ]
ospf area network  area network /mask  [restrict]
no ospf area network  area network /mask  [restrict]
[設定値及び初期値 ]
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値 非バックボーンエリア
IP アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : -
•network458 | コマンドリファレンス  | OSPF

• [設定値 ] : IP アドレス
• [初期値 ] : -
•mask
• [設定値 ] : ネットマスク長
• [初期値 ] : -
[説明 ]
エリア境界ルーターが他のエリアに経路を広告する場合に、 network/mask  で指定したネットワーク範囲内の個々の経
路を  network/mask  に要約して広告する。 restrict キーワードを指定した場合は、 network/mask  の範囲内の経路は要約
した経路も含めて一切他のエリアに広告しなくなる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.11 スタブ的接続の広告
[書式 ]
ospf area stubhost  area host [cost cost]
no ospf area stubhost  area host
[設定値及び初期値 ]
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値 非バックボーンエリア
IP アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : -
•host
• [設定値 ] : IP アドレス
• [初期値 ] : -
•cost
• [設定値 ] : 1 以上の数値
• [初期値 ] : -
[説明 ]
指定したホストが指定したコストでスタブ的に接続されていることをエリア内に広告する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.12 仮想リンク設定
[書式 ]
ospf virtual-link  router_id  area [parameters ...]
no ospf virtual-link  router_id  [area [parameters ...]]
[設定値及び初期値 ]
•router_id
• [設定値 ] : 仮想リンクの相手のルーター ID
• [初期値 ] : -
•area
• [設定値 ] :
設定値 説明
1 以上の数値 非バックボーンエリア
IP アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリアコマンドリファレンス  | OSPF | 459

• [初期値 ] : -
•parameters
• [設定値 ] : NAME=V ALUE の列
• [初期値 ] :
• retransmit-interval = 5 秒
• transmit-delay = 1 秒
• hello-interval = 10 秒
• dead-interval = 40 秒
• authkey= なし
• md5key= なし
• md5-sequence-mode=second
[説明 ]
仮想リンクを設定する。仮想リンクは router_id  で指定したルーターに対して、 area で指定したエリアを経由して設
定される。 parameters  では、仮想リンクのパラメータが設定できる。パラメータは  NAME=V ALUE の形で指定され、
以下の種類がある。
NAME V ALUE 説明
retransmit-interval 秒数LSA を連続して送る場合の再送間隔
を秒単位で設定する。  (1..)
transmit-delay 秒数リンクの状態が変わってから  LSA を
送信するまでの時間を秒単位で設定
する。  (1..)
hello-interval 秒数HELLO パケットの送信間隔を秒単
位で設定する。  (1..)
dead-interval 秒数相手から  HELLO を受け取れない場
合に、相手がダウンしたと判断する
までの時間を秒単位で設定する。
(1..)
authkey 文字列プレーンテキスト認証の認証鍵を表
す文字列を設定する。 （ 8 文字以内）
md5key "(ID),(KEY)"MD5 認証の認証鍵を表す  ID と鍵文
字列  KEY を設定する。 ID は十進数
で 0～255、KEY は文字列で  16 文字
以内。 MD5 認証鍵は  2 つまで設定で
きる。複数の  MD5 認証鍵が設定さ
れている場合には、送信パケットは
同じ内容のパケットを複数個、それ
ぞれの鍵による認証データを付加し
て送信する。受信時には鍵  ID が一
致する鍵が比較対象となる。
md5-sequence-mode"second" 送信時刻の秒数
"increment" 単調増加
[ノート ]
・hello-interval/dead-interval について
hello-interval と dead-interval の値は、そのインタフェースから直接通信できるすべての近隣ルーターとの間で同じ値
でなくてはいけない。これらのパラメータの値が設定値とは異なっている  OSPFHELLO パケットを受信した場合に
は、それは無視される。
・MD5 認証鍵について
MD5 認証鍵を複数設定できる機能は、 MD5 認証鍵を円滑に変更するためである。
通常の運用では、 MD5 認証鍵は  1 つだけ設定しておく。 MD5 認証鍵を変更する場合は、まず  1 つのルーターで新旧
の MD5 認証鍵を  2 つ設定し、その後、近隣ルーターで  MD5 認証鍵を新しいものに変更していく。そして、最後に
2 つの鍵を設定したルーターで古い鍵を削除すれば良い。460 | コマンドリファレンス  | OSPF

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.13 指定インタフェースの  OSPF エリア設定
[書式 ]
ip interface  ospf area area [parameters ...]
ip pp ospf area area [parameters ...]
ip tunnel  ospf area area [parameters ...]
no ip interface  ospf area [area [parameters ...]]
no ip pp ospf area [area [parameters ...]]
no ip tunnel  ospf area [area [parameters ...]]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 LOOPBACK インタフェース名
• [初期値 ] : -
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値 非バックボーンエリア
IP アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : インタフェースは  OSPF エリアに属していない
•parameters
• [設定値 ] : NAME=V ALUE の列
• [初期値 ] :
• type=broadcast(LAN インタフェース設定時  )
• type=point-to-point(PP 、 TUNNEL または  LOOPBACK インタフェース設定時  )
• passive= インタフェースは  passive ではない
• cost=1(LAN インタフェース、 LOOPBACK インタフェース設定時  )、pp は回線速度に依存
• priority=1
• retransmit-interval=5 秒
• transmit-delay=1 秒
• hello-interval=10 秒 (type=broadcast 設定時  )
• hello-interval=10 秒 (point-to-point 設定時  )
• hello-interval=30 秒 (non-broadcast 設定時  )
• hello-interval=30 秒 (point-to-multipoint 設定時  )
• dead-interval=hello-interval の 4 倍
• poll-interval=120 秒
• authkey= なし
• md5key= なし
• md5-sequence-mode=second
[説明 ]
指定したインタフェースの属する  OSPF エリアを設定する。
NAME パラメータの  type はインタフェースのネットワークがどのようなタイプであるかを設定する。
parameters  では、リンクパラメータを設定する。パラメータは  NAME=V ALUE の形で指定され、以下の種類があ
る。コマンドリファレンス  | OSPF | 461

NAME V ALUE 説明
typebroadcast ブロードキャスト
point-to-point ポイント・ポイント
point-to-multipoint ポイント・マルチポイント
non-broadcast NBMA
passiveインタフェースに対して、 OSPF パケ
ットを送信しない。該当インタフェ
ースに他の  OSPF ルーターがいない
場合に設定する。
cost コストインタフェースのコストを設定す
る。初期値は、インタフェースの種
類と回線速度によって決定される。
LAN インタフェースの場合は  1、PP
インタフェースの場合は、バインド
されている回線の回線速度を
S[kbit/s]とすると、以下の計算式で決
定される。例えば、 64kbit/s の場合は
1562、1.536Mbit/s の場合には  65 とな
る。 (0..65535)
• COST=100000/S
TUNNEL インタフェースの場合は、
1562 がデフォルト値となる。
priority 優先度指定ルーターの選択の際の優先度を
設定する。 PRIORITY 値が大きいル
ーターが指定ルーターに選ばれる。
0 を設定すると、 指定ルーターに選ば
れなくなる。 (0..255)
retransmit-interval 秒数LSA を連続して送る場合の再送間隔
を秒単位で設定する。  (1..)
transmit-delay 秒数リンクの状態が変わってから  LSA を
送信するまでの時間を秒単位で設定
する。  (1..)
hello-interval 秒数HELLO パケットの送信間隔を秒単
位で設定する。  (1..)
dead-interval 秒数近隣ルーターから  HELLO を受け取
れない場合に、近隣ルーターがダウ
ンしたと判断するまでの時間を秒単
位で設定する。  (1..)
poll-interval 秒数非ブロードキャストリンクでのみ有
効なパラメータで、近隣ルーターが
ダウンしている場合の  HELLO パケ
ットの送信間隔を秒単位で設定す
る。  (1..)
authkey 文字列プレーンテキスト認証の認証鍵を表
す文字列を設定する。 （ 8 文字以内）462 | コマンドリファレンス  | OSPF

NAME V ALUE 説明
md5key "(ID),(KEY)"MD5 認証の認証鍵を表す  ID と鍵文
字列  KEY を設定する。 ID は十進数
で 0～255、KEY は文字列で  16 文字
以内。 MD5 認証鍵は  2 つまで設定で
きる。複数の  MD5 認証鍵が設定さ
れている場合には、送信パケットは
同じ内容のパケットを複数個、それ
ぞれの鍵による認証データを付加し
て送信する。受信時には鍵  ID が一
致する鍵が比較対象となる。
md5-sequence-mode"second" 送信時刻の秒数
"increment" 単調増加
LOOPBACK インタフェースに設定する場合は、 type パラメータでインタフェースタイプを、 cost パラメータでイン
タフェースのコストを指定できる。 LOOPBACK インタフェースのタイプで指定できるのは、 以下の  2 種類だけとな
る。
NAME V ALUE広告される経路の種
類OSPF 的なインタフ
ェースの扱い
タイプ 状態
typeloopbackLOOPBACK インタ
フェースの  IP アドレ
スのみのホスト経路point-to-point Loopback
loopback-networkLOOPBACK インタ
フェースの  implicit
なネットワーク経路NBMA DROther
[ノート ]
・NAME パラメータの  type について
NAME パラメータの  type として、 LAN インタフェースは  broadcast のみが許される。 PP インタフェースは、 PPP を
利用する場合は  point-to-point 、フレームリレーを利用する場合は  point-to-multipoint と non-broadcast のいずれかが設
定できる。
フレームリレーで  non-broadcast(NBMA) を利用する場合には、フレームリレーの各拠点間のすべての間で  PVC が設
定されており、 FR に接続された各ルーターは他のルーターと直接通信できるような状態、すなわちフルメッシュに
なっていなくてはならない。また、 non-broadcast では近隣ルーターを自動的に認識することができないため、すべ
ての近隣ルーターを ip pp ospf neighbor  コマンドで設定する必要がある。
point-to-multipoint を利用する場合には、フレームリレーの  PVC はフルメッシュである必要はなく、一部が欠けたパ
ーシャルメッシュでも利用できる。近隣ルーターは  InArp を利用して自動的に認識するため、 InArp が必須となる。
RT では  InArp を使うかどうかは  fr inarp  コマンドで制御できるが、デフォルトでは  InArp を使用する設定になって
いるので、 ip pp address  コマンドでインタフェースに適切な  IP アドレスを与えるだけでよい。
point-to-multipoint と設定されたインタフェースでは、 ip pp ospf neighbor  コマンドの設定は無視される。
point-to-multipoint の方が  non-broadcast よりもネットワークの制約が少なく、また設定も簡単だが、その代わりに回
線を流れるトラフィックは大きくなる。 non-broadcast では、 broadcast と同じように指定ルーターが選定され、 HELLO
などの  OSPF トラフィックは各ルーターと指定ルーターの間だけに限定されるが、 point-tomultipoint ではすべての通
信可能なルーターペアの間に  point-to-point リンクがあるという考え方なので、 OSPF トラフィックもすべての通信可
能なルーターペアの間でやりとりされる。
・passive について
passive は、インタフェースが接続しているネットワークに他の  OSPF ルーターが存在しない場合に指定する。
passive を指定しておくと、インタフェースから  OSPF パケットを送信しなくなるので、無駄なトラフィックを抑制
したり、受信側で誤動作の原因になるのを防ぐことができる。
LAN インタフェース  (type=broadcast であるインタフェース  ) の場合には、インタフェースが接続しているネットワ
ークへの経路は、 ip interface  ospf area  コマンドを設定していないと他の  OSPF ルーターに広告されない。そのため、
OSPF を利用しないネットワークに接続する  LAN インタフェースに対しては、 passive を付けた ip interface  ospf areaコマンドリファレンス  | OSPF | 463

コマンドを設定しておくことでそのネットワークでは  OSPF を利用しないまま、 そこへの経路を他の  OSPF ルーター
に広告することができる。
PP インタフェースに対して ip interface  ospf area  コマンドを設定していない場合は、 インタフェースが接続するネッ
トワークへの経路は外部経路として扱われる。外部経路なので、他の  OSPF ルーターに広告するには ospf import  コ
マンドの設定が必要である。
・hello-interval/dead-interval について
hello-interval/dead-interval の値は、そのインタフェースから直接通信できるすべての近隣ルーターとの間で同じ値で
なくてはいけない。これらのパラメータの値が設定値とは異なっている  OSPF HELLO パケットを受信した場合に
は、それは無視される。
・MD5 認証鍵について
MD5 認証鍵を複数設定できる機能は、 MD5 認証鍵を円滑に変更するためである。
通常の運用では、 MD5 認証鍵は  1 つだけ設定しておく。 MD5 認証鍵を変更する場合は、まず  1 つのルーターで新旧
の MD5 認証鍵を  2 つ設定し、その後、近隣ルーターで  MD5 認証鍵を新しいものに変更していく。そして、最後に
2 つの鍵を設定したルーターで古い鍵を削除すれば良い。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.14 非ブロードキャスト型ネットワークに接続されている  OSPF ルーターの指定
[書式 ]
ip interface  ospf neighbor  ip_address  [eligible]
ip pp ospf neighbor  ip_address  [eligible]
ip tunnel  ospf neighbor  ip_address  [eligible]
no ip interface  ospf neighbor  ip_address  [eligible]
no ip pp ospf neighbor  ip_address  [eligible]
no ip tunnel  ospf neighbor  ip_address  [eligible]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•ip_address
• [設定値 ] : 近隣ルーターの  IP アドレス
• [初期値 ] : -
[説明 ]
非ブロードキャスト型のネットワークに接続されている  OSPF ルーターを指定する。
eligible キーワードが指定されたルーターは指定ルーターとして適格であることを表す。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.15 スタブが存在する時のネットワーク経路の扱いの設定
[書式 ]
ospf merge  equal  cost stub merge
no ospf merge  equal  cost stub
[設定値及び初期値 ]
•merge
• [設定値 ] :
設定値 説明
on イコールコストになるスタブを他の経路とマージす
る464 | コマンドリファレンス  | OSPF

設定値 説明
off イコールコストになるスタブを他の経路とマージし
ない
• [初期値 ] : on
[説明 ]
他の経路と同じコストになるスタブをどう扱うかを設定する。
on の場合にはスタブへの経路を他の経路とマージして、イコールコストマルチパス動作をする。これは、 RFC2328
の記述に沿うものである。
off の場合にはスタブへの経路を無視する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.16 OSPF の状態遷移とパケットの送受信をログに記録するか否かの設定
[書式 ]
ospf log log [log...]
no ospf log [log...]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
interface インタフェースの状態遷移
neighbor 近隣ルーターの状態遷移
packet 送受信したパケット
• [初期値 ] : OSPF のログは記録しない。
[説明 ]
指定した種類のログを  INFO レベルで記録する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
28.17 インタフェースの状態変化時、 OSPF　に外部経路を反映させる時間間隔の設定
[書式 ]
ospf reric  interval  time
no ospf reric  interval  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (1 以上の数値  )
• [初期値 ] : 1
[説明 ]
ルーターのインタフェースの状態が変化したとき、 OSPF に外部経路を反映させる時間の間隔を設定する。
OSPF ではインタフェースの状態変化を 1秒間隔で監視し、変化があれば最新の外部経路を自身に反映させるが、イ
ンタフェースの状態変化が連続して発生するときは、複数の外部経路の反映処理が  time で指定した秒数の間隔でま
とめて行われるようになる。コマンドリファレンス  | OSPF | 465

[ノート ]
複数のトンネルが一斉にアップすることがあるような環境では、本コマンドの値を適切に設定することで、 OSPF や
BGP の外部経路の導入によるシステムへの負荷を軽減することができる。
本コマンドの設定値は、 BGP への外部経路の反映にも影響する。本コマンドと  bgp reric interval  コマンドの設定値
が食い違う場合には、本コマンドの設定値が優先して適用される。
本コマンドの設定は、 経路の変化や  IP アドレスの変化に対する  OSPF や BGP の動作には関係しない。また本コマン
ドの設定値は、 ospf configure refresh  コマンドを実行しなくても即時反映される。
RTX1210 は Rev.14.01.16 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.21 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830466 | コマンドリファレンス  | OSPF

