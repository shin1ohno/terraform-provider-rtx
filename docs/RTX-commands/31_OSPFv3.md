# 第31章: OSPFv3

> 元PDFページ: 517-526

---

第 31 章
OSPFv3
31.1 OSPFv3 の有効設定
[書式 ]
ipv6 ospf configure  refresh
[説明 ]
OSPFv3 の設定を有効にする。 OSPFv3 関係の設定を変更したら、ルーターを再起動するか、あるいはこのコマンド
を実行しなくてはならない。
[ノート ]
このコマンドを入力したとき、次のいずれかならば、 OSPFv3 の設定は有効にならない。
•ルーター ID が設定されていない
•エリアが設定されていない
•いずれのインタフェースもエリアに属していない
•仮想リンクが経由するエリアが存在しない
•仮想リンクが経由するエリアに属するインタフェースが存在しない
すでに  OSPFv3 の設定が有効であるときにこのコマンドを入力した場合、初期状態から再度設定を読み込む。
よって、それまで  OSPFv3 が保持していた経路情報や、他のプロトコルに配布した経路情報は一旦破棄され、初期
状態から動作を開始する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.2 OSPFv3 の使用設定
[書式 ]
ipv6 ospf use use
no ipv6 ospf use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on OSPFv3 を使用する
off OSPFv3 を使用しない
• [初期値 ] : off
[説明 ]
OSPFv3 を使用するか否かを設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.3 OSPFv3 のルーター ID 設定
[書式 ]
ipv6 ospf router  id router-id
no ipv6 ospf router  id [router-id ]
[設定値及び初期値 ]
•router_id
• [設定値 ] : IPv4 アドレス表記  (0.0.0.0 は不可  )コマンドリファレンス  | OSPFv3 | 517

• [初期値 ] : -
[説明 ]
ルーター ID を設定する。
[ノート ]
ipv6 ospf configure refresh  コマンドが入力されたとき、 このコマンドによってルーター ID が設定されていない場合、
以下の順序でインタフェースに付与されているプライマリ IPv4アドレスを探索して  最初に見つかった IPv4アドレ
スをルーター IDとして使用する。
• LANインタフェース (若番順 )
• LOOPBACK インタフェース (若番順）
プライマリ  IPv4 アドレスが付与されたインタフェースがない場合は初期値は設定されない。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.4 OSPFv3 エリア設定
[書式 ]
ipv6 ospf area area [stub [cost= cost]]
no ipv6 ospf area area [stub [cost= cost]]
[設定値及び初期値 ]
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値  (1...4294967295) 非バックボーンエリア
IPv4 アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : -
•cost
• [設定値 ] : デフォルト経路のコスト  (0～16777215)
• [初期値 ] : 0
[説明 ]
OSPFv3 エリアを設定する。
stub キーワードを指定した場合、そのエリアはスタブエリアであることを表わす。 cost は 0 以上の数値で、エリア境
界ルーターがエリア内に広告するデフォルト経路のコストとして使われる。 cost を指定しないとデフォルト経路の
広告は行われない。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.5 エリアへの経路広告
[書式 ]
ipv6 ospf area network  area ipv6_prefix /prefix_len  [restrict]
no ipv6 ospf area network  area ipv6_prefix /prefix_len  [restrict]
[設定値及び初期値 ]
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値  (1...4294967295) 非バックボーンエリア
IPv4 アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア518 | コマンドリファレンス  | OSPFv3

• [初期値 ] : -
•ipv6_prefix/prefix_len
• [設定値 ] : IPv6 プレフィクス
• [初期値 ] : サブネットの範囲は設定されていない
[説明 ]
エリア境界ルーターが他のエリアに経路を広告する場合に、このコマンドで指定したサブネットの範囲内の経路は
単一のサブネット経路として広告する。 restrict キーワードが指定された場合には、範囲内の経路は要約した経路も
広告しない。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.6 指定インタフェースの  OSPFv3 エリア設定
[書式 ]
ipv6 interface  ospf area area [parameters  ...]
ipv6 pp ospf area area [parameters ...]
ipv6 tunnel  ospf area area [parameters ...]
no ipv6 interface  ospf area [area [parameters ...]]
no ipv6 pp ospf area [area [parameters ...]]
no ipv6 tunnel  ospf area [area [parameters ...]]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•area
• [設定値 ] :
設定値 説明
backbone バックボーンエリア
1 以上の数値  (1...4294967295) 非バックボーンエリア
IPv4 アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : インタフェースは  OSPF エリアに属していない
•parameters
• [設定値 ] : NAME=V ALUE の列
• [初期値 ] :
• type=broadcast(LAN インタフェース設定時  )
• type=point-to-point(PP インタフェース設定時、トンネルインタフェース設定時  )
• passive= インタフェースは  passive ではない
• cost=1(LAN インタフェース設定時  )、1562( トンネル設定時  )、pp は回線速度に依存
• priority=1
• retransmit-interval=5 秒
• transmit-delay=1 秒
• hello-interval=10 秒
• dead-interval=40 秒
[説明 ]
指定したインタフェースの属する  OSPFv3 エリアを設定する。 NAME パラメータの  type はインタフェースの接続す
るリンクがどのようなタイプであるかを設定する。 parameters  では、リンクパラメータを設定する。パラメータは
NAME=V ALUE の形で指定され、以下の種類がある。コマンドリファレンス  | OSPFv3 | 519

NAME V ALUE 説明
typebroadcast ブロードキャスト型
point-to-point ポイント・ポイント型
passiveインタフェースに対して、 OSPFv3 パ
ケットを送信しない。該当インタフ
ェースに他の  OSPFv3 ルーターがい
ない場合に設定する。
cost コスト  (1...65535)インタフェースのコストを設定す
る。初期値はインタフェースの種類
と回線速度によって決定される。
LAN インタフェースの場合は  1、ト
ンネルインタフェースの場合は
1562、PP インタフェースの場合は、
バインドされている回線の回線速度
を S[kbit/s]とすると、 以下の計算式で
決定される。例えば、 64kbit/s の場合
は 1562、1.536Mbit/s の場合には  65 と
なる。
cost=100000/S
priority 優先度  (0...255)指定ルーター選択の際の優先度を設
定する。値が大きいルーターが指名
ルーターに選ばれる。 0 を指定する
と指定ルーターに選ばれなくなる。
retransmit-interval 秒数  (1...65535)LSA を連続して送る場合の再送間隔
を秒単位で指定する。
transmit-delay 秒数リンクの状態が変わってから  LSA を
送信するまでの時間を秒単位で設定
する。
hello-interval 秒数  (1...65535)HELLO パケットの送信間隔を秒単
位で設定する
dead-interval 秒数  (1...65535)近隣ルーターから  HELLO を受け取
れない場合に、近隣ルーターがダウ
ンしたと判断するまでの時間を秒単
位で設定する。
[ノート ]
・NAME パラメータの  type について
NAME パラメータの  type として、 LAN インタフェースは  broadcast のみが設定できる。 PP インタフェースで  PPP を
利用する場合や、トンネルインタフェースを利用する場合は、 point-to-point のみが設定できる。
・passive について
passive は、インタフェースが接続しているリンクに他の  OSPFv3 ルーターが存在しない場合に指定する。 passive を
指定しておくと、インタフェースから  OSPFv3 パケットを送信しなくなるので、無駄なトラフィックを抑制したり、
受信側で誤動作の原因になるのを防ぐことができる。
LAN インタフェース  (type=broadcast であるインタフェース  ) の場合には、インタフェースが接続しているリンクへ
の経路は、 ipv6 interface  ospf area  コマンドを設定していないと他の  OSPFv3 ルーターに広告されない。そのため、
OSPFv3 を利用しないリンクに接続する  LAN インタフェースに対しては、 passive を付けた ipv6 interface  ospf area  コ
マンドを設定しておくことでそのリンクでは  OSPFv3 を利用しないまま、 そこへの経路を他の  OSPFv3 ルーターに広
告することができる。
PP インタフェースに対して ipv6 pp ospf area  コマンドを設定していない場合は、インタフェースが接続するリンク
への経路は外部経路として扱われる。外部経路なので、他の  OSPFv3 ルーターに広告するには ipv6 ospf import  コマ
ンドの設定が必要である。520 | コマンドリファレンス  | OSPFv3

・hello-interval/dead-interval について
hello-interval/dead-interval の値は、そのインタフェースから直接通信できるすべての近隣ルーターとの間で同じ値で
なくてはならない。これらのパラメータが設定値とは異なる  HELLO パケットを受信した場合には、 それは無視され
る。 dead-interval を指定しなかった場合には、 hello-interval の4倍の値が設定される。
・インスタンス  ID について
本機のインスタンス  ID は常に  0 であり、 OSPFv3 パケットを受信する場合には、同じ値を持つパケットのみを受信
する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.7 仮想リンク設定
[書式 ]
ipv6 ospf virtual-link  router_id  area [parameters  ...]
no ipv6 ospf virtual-link  router_id  [area [parameters ...]]
[設定値及び初期値 ]
•router_id
• [設定値 ] : 仮想リンクの相手のルーター ID
• [初期値 ] : -
•area : 経由するエリア
• [設定値 ] :
設定値 説明
1 以上の数値  (1...4294967295) 非バックボーンエリア
IPv4 アドレス表記  (0.0.0.0 は不可  ) 非バックボーンエリア
• [初期値 ] : -
•parameters
• [設定値 ] : NAME=V ALUE の列
• [初期値 ] :
• retransmit-interval=5 秒
• transmit-delay=1 秒
• hello-interval=10 秒
• dead-interval=40 秒
[説明 ]
仮想リンクを設定する。仮想リンクは router_id  で指定したルーターに対して、 area で指定したエリアを経由して設
定される。 parameters  では、仮想リンクのパラメータが設定できる。パラメータは  NAME=V ALUE の形で指定され、
以下の種類がある。
NAME V ALUE 説明
retransmit-interval 秒数  (1...65535)LSA を連続して送る場合の再送間隔
を秒数で設定する。
transmit-delay 秒数  (1...65535)リンクの状態が変わってから  LSA を
送信するまでの時間を秒単位で設定
する。
hello-interval 秒数  (1...65535)HELLO パケットの送信間隔を秒単
位で設定する。
dead-interval 秒数  (1...65535)相手から  HELLO を受け取れない場
合に、相手がダウンしたと判断する
までの時間を秒単位で設定する。コマンドリファレンス  | OSPFv3 | 521

[ノート ]
・hello-interval/dead-interval について
hello-interval と dead-interval の値は、そのインタフェースから直接通信できるすべての近隣ルーターとの間で同じ値
でなくてはならない。これらのパラメータの値が設定値とは異なっている  HELLO パケットを受信した場合には、 そ
れは無視される。
・インスタンス  ID について
本機のインスタンス  ID は常に  0 であり、 OSPFv3 パケットを受信する場合には、同じ値を持つパケットのみを受信
する。
・出力インタフェースについて
仮想リンクを設定した場合、経由するエリアへの出力インタフェースにグローバルアドレスが付与されていなけれ
ば、仮想リンクは使用できない。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.8 OSPFv3 による経路の優先度設定
[書式 ]
ipv6 ospf preference  preference
no ipv6 ospf preference  [preference ]
[設定値及び初期値 ]
•preference
• [設定値 ] : OSPFv3 による経路の優先度  (1...2147483647)
• [初期値 ] : 2000
[説明 ]
OSPFv3 による経路の優先度を設定する。優先度は  1 以上の数値で表され、数字が大きい程優先度が高い。
OSPFv3 と RIPng など複数のプロトコルで得られた経路が食い違う場合には、優先度が高い方が採用される。
優先度が同じ場合には時間的に先に採用された経路が有効となる。
[ノート ]
静的経路の優先度は  10000 で固定である。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.9 OSPFv3 で受け取った経路をルーティングテーブルに反映させるか否かの設定
[書式 ]
ipv6 ospf export  from  ospf filter filter_num  ...
no ipv6 ospf export  from  ospf [filter filter_num ...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : ipv6 ospf export filter  コマンドのフィルタ番号  (1...2147483647)
• [初期値 ] : すべての経路がルーティングテーブルに反映される
[説明 ]
OSPFv3 で受け取った経路をルーティングテーブルに導入するかどうかを設定する。指定した順にフィルタを評価
し、最初に合致したフィルタによって導入すると判断された経路だけがルーティングテーブルに導入される。
導入しないと判断された経路や合致するフィルタがない経路は導入されない。
このコマンドが設定されていない場合には、すべての経路がルーティングテーブルに導入される。522 | コマンドリファレンス  | OSPFv3

[ノート ]
このコマンドは  OSPFv3 のリンク状態データベースには影響を与えない。つまり、 OSPFv3 で他のルーターと情報を
やり取りする動作としては、 このコマンドがどのように設定されていても変化はない。 OSPFv3 で計算した経路が実
際にパケットをルーティングするために使われるかどうかだけが変わる。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.10 OSPFv3 で受け取った経路をどう扱うかのフィルタの設定
[書式 ]
ipv6 ospf export  filter filter_num  [nr] kind ipv6_prefix /prefix_len  ...
no ipv6 ospf export  filter filter_num [...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号  (1...2147483647)
• [初期値 ] : -
•nr : フィルタの解釈の方法
• [設定値 ] :
設定値 説明
not IPv6 プレフィクスに該当しない経路を導入する
reject IPv6 プレフィクスに該当した経路を導入しない
• [初期値 ] : -
•kind : IPv6 プレフィクスの解釈の方法
• [設定値 ] :
設定値 説明
include 指定した  IPv6 プレフィクスに含まれる経路  (IPv6 プ
レフィクス自身を含む  )
refines 指定した  IPv6 プレフィクスに含まれる経路  (IPv6 プ
レフィクス自身を含まない  )
equal 指定した  IPv6 プレフィクスに一致する経路
• [初期値 ] : -
•ipv6_prefix/prefix_len
• [設定値 ] : IPv6 プレフィクス
• [初期値 ] : -
[説明 ]
OSPFv3 により他の  OSPFv3 ルーターから受け取った経路をルーティングテーブルに導入する際に適用するフィル
タを定義する。このコマンドで定義したフィルタは、 ipv6 ospf export from ospf  コマンドの filter  項で指定されては
じめて効果を持つ。
ipv6_prefix/prefix_len  では、 IPv6 プレフィクスを設定する。これは複数設定でき、 kind に指定した方法で解釈される。
includeIPv6 プレフィクスと一致する経路および、 IPv6 プレフィ
クスに含まれる経路が該当
refinesIPv6 プレフィクスに含まれる経路が該当するが、 IPv6 プ
レフィクスと一致する経路は含まれない
equal IPv6 プレフィクスに一致する経路のみ該当
nr が省略されている場合には、一つでも該当する  IPv6 プレフィクスがある場合にフィルタに合致したものとし、そ
の経路を導入する。 not 指定時には、いずれの  IPv6 プレフィクスにも該当しなかった場合にフィルタに合致したも
のとし、その経路を導入する。 reject 指定時には、一つでも該当する  IPv6 プレフィクスがある場合にフィルタに合
致したものとし、その経路を導入しない。コマンドリファレンス  | OSPFv3 | 523

[ノート ]
not 指定のフィルタを ipv6 ospf export from ospf  コマンドで複数設定する場合には注意が必要である。 not 指定のフ
ィルタに合致する  IPv6 プレフィクスは、そのフィルタでは導入するかどうかが決定しないため、 ipv6 ospf export
from ospf  コマンドで指定された次のフィルタで評価される。そのため、例えば、以下のような設定ではすべての経
路が導入されることになりフィルタの意味がない。
ipv6 ospf export from ospf filter 1 2
ipv6 ospf export filter 1 not equal fec0:12ab:34cd:1::/64
ipv6 ospf export filter 2 not equal fec0:12ab:34cd:2::/64
1 番のフィルタは  fec0:12ab:34cd:1::/64 以外の経路に合致し、 2 番のフィルタは  fec0:12ab:34cd:2::/64 以外の経路に合
致する。つまり、経路  fec0:12ab:34cd:1::/64 は 1 番のフィルタに合致しないが、 2 番のフィルタに合致するため導入
される。一方で経路  fec0:12ab:34cd:2::/64 は 1 番のフィルタに合致するため、 2 番のフィルタにかかわらず導入され
る。よって、導入されない経路は存在しない。
経路  fec0:12ab:34cd:1::/64 と経路  fec0:12ab:34cd:2::/64 を導入したくない場合には以下のような設定を行なう必要が
ある
ipv6 ospf export from ospf filter 1
ipv6 ospf export filter 1 not equal fec0:12ab:34cd:1::/64 fec0:12ab:34cd:2::/64
あるいは
ipv6 ospf export from ospf filter 1 2 3
ipv6 ospf export filter 1 reject equal fec0:12ab:34cd:1::/64
ipv6 ospf export filter 2 reject equal fec0:12ab:34cd:2::/64
ipv6 ospf export filter 3 include ::/0
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.11 外部プロトコルによる経路導入
[書式 ]
ipv6 ospf import  from  protocol  [filter filter_num  ...]
no ipv6 ospf import  from  protocol  [filter filter_num ...]
[設定値及び初期値 ]
•protocol  : OSPFv3 の経路テーブルに導入する外部プロトコル
• [設定値 ] :
設定値 説明
static 静的経路  ( implicit 経路は含まない  )
rip RIPng
• [初期値 ] : -
•filter_num
• [設定値 ] : ipv6 ospf import filter  コマンドのフィルタ番号  (1...2147483647)
• [初期値 ] : -
[説明 ]
OSPFv3 の経路テーブルに外部プロトコルによる経路を導入するかどうかを設定する。導入した経路は外部経路と
して他の  OSPFv3 ルーターに広告される。
filter_num  はipv6 ospf import filter  コマンドで定義したフィルタ番号を指定する。外部プロトコルから導入されよう
とする経路は指定した順にフィルタにより評価される。最初に合致したフィルタによって導入すると判断された経
路は  OSPFv3 に導入される。導入しないと判断された経路や合致するフィルタがない経路は導入されない。また、
filter キーワード以降を省略した場合には、すべての経路が  OSPFv3 に導入される
経路を広告する場合のパラメータであるメトリック値、メトリックタイプは、フィルタの検査で該当した ipv6 ospf
import filter  コマンドで指定されたものを使う。 filter キーワード以降を省略した場合には、 以下のパラメータを使用
する。
• metric=1
• type=2524 | コマンドリファレンス  | OSPFv3

[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.12 外部経路導入に適用するフィルタ定義
[書式 ]
ipv6 ospf import  filter filter_num  [nr] kind ipv6_prefix /prefix_len  ... [parameters  ...]
no ipv6 ospf import  filter filter_num  [[nr] kind ipv6_prefix /prefix_len  ... [parameters ...]]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号  (1...2147483647)
• [初期値 ] : -
•nr : フィルタの解釈の方法
• [設定値 ] :
設定値 説明
not IPv6 プレフィクスに該当しない経路を導入する
reject IPv6 プレフィクスに該当する経路を導入しない
• [初期値 ] : -
•kind : IPv6 プレフィクスの解釈の方法
• [設定値 ] :
設定値 説明
include 指定した  IPv6 プレフィクスに含まれる経路  (IPv6 プ
レフィクス自身を含む  )
refines 指定した  IPv6 プレフィクスに含まれる経路  (IPv6 プ
レフィクス自身を含まない  )
equal 指定した  IPv6 プレフィクスに一致する経路
• [初期値 ] : -
•ipv6_prefix/prefix_len
• [設定値 ] : IPv6 プレフィクス
• [初期値 ] : -
•parameters  : 外部経路を広告する場合のパラメータ
• [設定値 ] :
設定値 説明
metric = metric メトリック値  (1 ～ 16777215)
type = type メトリックのタイプ  (1 または  2)
• [初期値 ] : -
[説明 ]
OSPFv3 の経路テーブルに外部経路を導入する際に適用するフィルタを定義する。このコマンドで定義したフィル
タは ipv6 ospf import from  コマンドの  filter 項で指定されてはじめて効果を持つ。
ipv6_prefix/prefix_len  では  IPv6 プレフィクスを指定する。これは複数指定でき、 kind に指定した方法で解釈される。
includeIPv6 プレフィクスと一致する経路および、 IPv6 プレフィ
クスに含まれる経路が該当
refinesIPv6 プレフィクスに含まれる経路が該当するが、 IPv6 プ
レフィクスと一致する経路は該当しない
equal IPv6 プレフィクスに一致する経路のみ該当
nr が省略されている場合には、一つでも該当する  IPv6 プレフィクスがある場合にフィルタに合致したものとし、そ
の経路を導入する。 not 指定時には、いずれの  IPv6 プレフィクスにも該当しなかった場合にフィルタに合致したもコマンドリファレンス  | OSPFv3 | 525

のとし、その経路を導入する。 reject 指定時には、一つでも該当する  IPv6 プレフィクスがある場合にフィルタに合
致したものとし、その経路を導入しない。
parameters  では、導入する経路を  OSPFv3 の外部経路として広告する場合のパラメータとして、メトリック値、メト
リックタイプがそれぞれ  metric、type により指定できる。これらを省略した場合には、以下の値が採用される。
• metric=1
• type=2
[ノート ]
not 指定のフィルタを ipv6 ospf import from  コマンドで複数設定する場合には注意が必要である。 not 指定のフィル
タに合致しない経路は、そのフィルタでは導入するかどうかが決定しないため、 ipv6 ospf import from  コマンドで指
定された次のフィルタで評価される。そのため、例えば以下のような設定ではすべての経路が導入されることにな
りフィルタの意味がない。
ipv6 ospf import from static filter 1 2
ipv6 ospf import filter 1 not equal fec0:12ab:34cd:1::/64
ipv6 ospf import filter 2 not equal fec0:12ab:34cd:2::/64
1 番のフィルタは  fec0:12ab:34cd:1::/64 以外の経路に合致し、 2 番のフィルタは  fec0:12ab:34cd:2::/64 以外の経路に合
致する。つまり、経路  fec0:12ab:34cd:1::/64 は 1 番のフィルタに合致しないが、 2 番のフィルタに合致するため導入
される。一方で経路  fec0:12ab:34cd:2::/64 は 1 番のフィルタに合致するため、 2 番のフィルタにかかわらず導入され
る。よって、導入されない経路は存在しない。
経路  fec0:12ab:34cd:1::/64 と経路  fec0:12ab:34cd:2::/64 を導入したくない場合には以下のような設定を行う必要があ
る。
ipv6 ospf import from static filter 1
ipv6 ospf import filter 1 not equal fec0:12ab:34cd:1::/64 fec0:12ab:34cd:2::/64
あるいは
ipv6 ospf import from static filter 1 2 3
ipv6 ospf import filter 1 reject equal fec0:12ab:34cd:1::/64
ipv6 ospf import filter 2 reject equal fec0:12ab:34cd:2::/64
ipv6 ospf import filter 3 include ::/0
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
31.13 OSPFv3 のログ出力設定
[書式 ]
ipv6 ospf log log ...
no ipv6 ospf log [log...]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
interface インタフェースの状態や仮想リンクに関わるログ
neighbor 近隣ルーターの状態に関わるログ
packet OSPFv3 パケットに関わるログ
• [初期値 ] : いずれの種類のログも出力しない
[説明 ]
OSPFv3 に関わるログ出力の種類を設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830526 | コマンドリファレンス  | OSPFv3

