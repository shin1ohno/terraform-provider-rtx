# 第29章: BGP

> 元PDFページ: 467-478

---

第 29 章
BGP
29.1 BGP の起動の設定
[書式 ]
bgp use use
no bgp use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 起動する
off 起動しない
• [初期値 ] : off
[説明 ]
BGP を起動するか否かを設定する
[ノート ]
いずれかのインタフェースにセカンダリアドレスを割り当てた場合、 BGP を使用することはできない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.2 経路の集約の設定
[書式 ]
bgp aggregate  ip_address /mask  filter filter_num  ...
no bgp aggregate  ip_address /mask  [filter filter_num ... ]
[設定値及び初期値 ]
•ip_address/mask
• [設定値 ] : IP アドレス /ネットマスク
• [初期値 ] : -
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -
[説明 ]
BGP で広告する集約経路を設定する。フィルタの番号には、 bgp aggregate filter  コマンドで定義した番号を指定す
る。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.3 経路を集約するためのフィルタの設定
[書式 ]
bgp aggregate  filter  filter_num  protocol  [reject] kind ip_address /mask  ...
no bgp aggregate  filter  filter_num  [protocol  [reject] kind ip_address /mask  ...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -コマンドリファレンス  | BGP | 467

•protocol
• [設定値 ] :
設定値 説明
static 静的経路  ( implicit 経路を含む  )
rip RIP
ospf OSPF
bgp BGP
all すべてのプロトコル
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
include 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含む  )
refines 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含まない  )
equal 指定したネットワークに一致する経路
• [初期値 ] : -
•ip_address/mask
• [設定値 ] : IP アドレス /ネットマスク
• [初期値 ] : -
[説明 ]
BGP で広告する経路を集約するためのフィルタを定義する。このコマンドで定義したフィルタは、 bgp aggregate  コ
マンドの  filter 節で指定されてはじめて効果を持つ。
ip_address/mask  では、ネットワークアドレスを設定する。これは複数設定でき、そのうち、一致するネットワーク
長が長い設定が採用される。
kind の前に  reject キーワードを置くと、その経路は集約されない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.4 AS 番号の設定
[書式 ]
bgp autonomous-system  as
no bgp autonomous-system  [as]
[設定値及び初期値 ]
•as
• [設定値 ] : AS 番号  (1..65535)
• [初期値 ] : -
[説明 ]
ルーターの  AS 番号を設定する。
[ノート ]
AS 番号を設定するまで  BGP は動作しない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830468 | コマンドリファレンス  | BGP

29.5 ルーター ID の設定
[書式 ]
bgp router  id ip_address
no bgp router  id [ip_address ]
[設定値及び初期値 ]
•ip_address
• [設定値 ] : IP アドレス
• [初期値 ] : インタフェースに付与されているプライマリアドレスから自動的に選択する。
[説明 ]
ルーター ID を設定する。
[ノート ]
ルーター  ID が本コマンドで設定されていないときは、以下のインタフェースに付与されているプライマリ  IPv4 ア
ドレスのいずれかが自動的に選択され、ルーター  ID として使用させれる。
• LAN インタフェース
• LOOPBACK インタフェース
• PP インタフェース
なお、プライマリ  IPv4 アドレスが付与されたインタフェースがない場合は初期値は設定されない。
意図しない  IP アドレスがルーター  ID として使用されることを防ぐため、本コマンドにより明示的にルーター  ID を
指定することが望ましい。
OSPF と BGP-4 とを併用する場合、本コマンドか  ospf router id コマンドのいずれか一方を設定する。
以下のリビジョンでは、本コマンドと  ospf router id コマンドの両方を設定することができるが、必ず同一のルータ
ー ID を指定する必要がある。
機種 リビジョン
vRX シリーズ すべてのリビジョン
RTX3510 すべてのリビジョン
RTX1300 すべてのリビジョン
RTX1220 すべてのリビジョン
RTX1210 Rev.14.01.16
RTX5000 Rev.14.00.21
RTX3500 Rev.14.00.21
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.6 BGP による経路の優先度の設定
[書式 ]
bgp preference  preference
no bgp preference  [preference ]
[設定値及び初期値 ]
•preference
• [設定値 ] : 優先度  (1..2147483647)
• [初期値 ] : 500
[説明 ]
BGP による経路の優先度を設定する。優先度は  1 以上の整数で示され、数字が大きいほど優先度が高い。 BGP とそ
の他のプロトコルで得られた経路が食い違う場合には、優先度の高い経路が採用される。優先度が同じ場合には、
先に採用された経路が有効になる。コマンドリファレンス  | BGP | 469

[ノート ]
各プロトコルに与えられた優先度の初期値は次のとおり。
スタティック 10000
RIP 1000
OSPF 2000
BGP 500
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.7 BGP で受信した経路に対するフィルタの適用
[書式 ]
bgp export  remote_as  filter filter_num  ...
bgp export  aspath  seq "aspath_regexp"  filter filter_num  ...
no bgp export  remote_as  [filter filter_num  ...]
no bgp export  aspath  seq ["aspath_regexp"  [filter filter_num  ...]]
[設定値及び初期値 ]
•remote_as
• [設定値 ] : 相手の  AS 番号  (1..65535)
• [初期値 ] : -
•seq
• [設定値 ] : AS パスを指定したときの評価順序  (1..65535)
• [初期値 ] : -
•aspath_regexp
• [設定値 ] : 正規表現
• [初期値 ] : -
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -
[説明 ]
BGP で受けた経路に対してフィルタを設定する。 remote_as  を指定してフィルタを設定した場合、接続先から受けた
経路についてフィルタに該当した経路が実際のルーティングテーブルに導入され、 RIP や OSPF のような他のプロト
コルにも通知される。フィルタに該当しない経路はルーティングには適用されず、他のプロトコルに通知されるこ
ともない。フィルタの番号には bgp export filter  コマンドで定義した番号を指定する。
aspath_regexp  を指定してフィルタを設定した場合、 remote_as  を指定した場合と同様に、 AS パスが正規表現と一致
する経路についてフィルタに該当した経路が導入される。 aspath_regexp  には grep  コマンドで使用できる検索パタ
ーンを指定する。
aspath_regexp  を指定したフィルタを複数設定した場合、 seq の小さい順に評価される。また、 aspath_regexp  を指定し
たフィルタを設定した場合、  remote_as  を指定したフィルタよりも優先して評価される。
[ノート ]
正規表現によって  AS パスを表す例
•すべての  AS パスと一致する
# bgp export aspath 10 ".*" filter 1
• AS 番号が  1000 または  1100 で始まる  AS パスと一致する
# bgp export aspath 20 "^1[01]00 .*" filter 1
• AS 番号に  2000 を含む  AS パスと一致する
# bgp export aspath 30 "2000" filter 1
• AS パスが  3000 3100 3200 であるパスと完全一致する470 | コマンドリファレンス  | BGP

# bgp export aspath 40 "^3000 3100 3200$" filter 1
• AS パスに  AS_SET を含むパスと一致する
# bgp export aspath 50 "{.*}" filter 1
フィルタ番号は、 101個まで設定できる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.8 BGP で受信する経路に適用するフィルタの設定
[書式 ]
bgp export  filter  filter_num  [reject] kind ip_address /mask  ... [parameter  ]
no bgp export  filter  filter_num  [[reject] kind ip_address /mask  ... [parameter ]]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
include 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含む  )
refines 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含まない  )
equal 指定したネットワークに一致する経路
• [初期値 ] : -
•ip_address/mask
• [設定値 ] :
設定値 説明
ip_address/mask IP アドレス /ネットマスク
all すべてのネットワーク
• [初期値 ] : -
•parameter  : TYPE=V ALUE の組
• [設定値 ] :
TYPE V ALUE 説明
preference 0..255同じ経路を複数の相手から受信し
たときに、一方を選択するための
優先度
• [初期値 ] : 0
[説明 ]
BGP で受信する経路に適用するフィルタを定義する。このコマンドで定義したフィルタは、 bgp export  コマンドの
filter 節で指定されてはじめて効果を持つ。
ip_address/mask  では、ネットワークアドレスを設定する。複数の設定があるときには、プレフィックスが最も長く
一致する設定が採用される。
kind の前に  reject キーワードを置くと、その経路が拒否される。
[ノート ]
preference の設定は  BGP 経路の間で優先順位をつけるために使用される。 BGP 経路の全体の優先度は、 bgp
preference  コマンドで設定する。コマンドリファレンス  | BGP | 471

[設定例 ]
# bgp export filter 1 include 10.0.0.0/16 172.16.0.0/16
# bgp export filter 2 reject equal 192.168.0.0/24
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.9 BGP に導入する経路に対するフィルタの適用
[書式 ]
bgp import  remote_as  protocol  [from_as ] filter filter_num  ...
no bgp import  remote_as  protocol  [from_as ] [filter filter_num  ...]
[設定値及び初期値 ]
•remote_as
• [設定値 ] : 相手の  AS 番号  (1..65535)
• [初期値 ] : -
•protocol
• [設定値 ] :
設定値 説明
static 静的経路  ( implicit 経路を含む  )
rip RIP
ospf OSPF
bgp BGP
aggregate 集約経路
• [初期値 ] : -
•from_as
• [設定値 ] : 導入する経路を受信した  AS(protocol  で bgp を指定したときのみ  )(1..65535)
• [初期値 ] : -
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -
[説明 ]
RIP や OSPF のような  BGP 以外の経路を導入するときに適用するフィルタを設定する。フィルタに該当しない経路
は導入されない。フィルタの番号には、 bgp import filter  コマンドで定義した番号を指定する。 BGP の経路を導入す
るときには、その経路を受信した  AS 番号を指定する必要がある。
[ノート ]
このコマンドが設定されていないときには、外部経路は導入されない。
フィルタ番号は、 102個まで設定できる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.10 BGP の設定の有効化
[書式 ]
bgp configure  refresh
[説明 ]
BGP の設定を有効にする。 BGP の設定を変更したら、ルーターを再起動するか、このコマンドを実行する必要があ
る。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830472 | コマンドリファレンス  | BGP

29.11 BGP に導入する経路に適用するフィルタの設定
[書式 ]
bgp import  filter  filter_num  [reject] kind ip_address /mask  ... [parameter  ...]
no bgp import  filter  filter_num  [[reject] kind ip_address /mask  ... [parameter  ...]]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : フィルタ番号  (1..2147483647)
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
include 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含む  )
refines 指定したネットワークに含まれる経路  ( ネットワー
クアドレス自身を含まない  )
equal 指定したネットワークに一致する経路
• [初期値 ] : -
•ip_address/mask
• [設定値 ] :
設定値 説明
ip_address/mask IP アドレス /ネットマスク
all すべてのネットワーク
• [初期値 ] : -
•parameter  : TYPE=V ALUE の組
• [設定値 ] :
TYPE V ALUE 説明
metric 1..16777215MED(Multi-Exit Discriminator) で通
知するメトリック値  ( 指定しない
ときは  MED を送信しない  )
preference 0..255 Local Preference で通知する優先度
• [初期値 ] :
• preference=100
[説明 ]
BGP に導入する経路に適用するフィルタを定義する。このコマンドで定義したフィルタは、 bgp import  コマンドの
filter 節で指定されてはじめて効果を持つ。
ip_address/mask  では、ネットワークアドレスを設定する。複数の設定があるときには、プレフィックスが最も長く
一致する設定が採用される。
kind の前に  reject キーワードを置くと、その経路が拒否される。
[設定例 ]
# bgp import filter 1 include 10.0.0.0/16 172.16.0.0/16
# bgp import filter 2 reject equal 192.168.0.0/24
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | BGP | 473

29.12 BGP による接続先の設定
[書式 ]
bgp neighbor  neighbor_id  remote_as  remote_address  [parameter ...]
no bgp neighbor  neighbor_id  [remote_as  remote_address  [parameter ...]]
[設定値及び初期値 ]
•neighbor_id
• [設定値 ] : 近隣ルーターの番号  (1...2147483647)
• [初期値 ] : -
•remote_as
• [設定値 ] : 相手の  AS 番号  (1..65535)
• [初期値 ] : -
•remote_address
• [設定値 ] : 相手の  IP アドレス
• [初期値 ] : -
•parameter  : TYPE=V ALUE の組
• [設定値 ] :
TYPE V ALUE 説明
hold-time off、秒数ダウン判定までの時間  (3..28,800
秒 )
metric 1..21474836MED(Multi-Exit Discriminator) で通
知するメトリック
passive on または  off能動的な  BGP コネクションの接続
を抑制するか否か
gateway IP アドレス /インタフェース 接続先に対するゲートウェイ
local-address IP アドレスBGP コネクションの自分のアドレ
ス
ignore-capability on または  off capability を無視するか否か
• [初期値 ] :
• hold-time=180
• metric は送信されない
• passive=off
• gateway は指定されない
• local-address は指定されない
• ignore-capability=off
[説明 ]
BGP コネクションを接続する近隣ルーターを定義する。
[ノート ]
hold-time パラメータに設定した時間の  1/3 程度の間隔で  KEEPALIVE メッセージを送信する。 hold-time パラメータ
の時間だけ待っても  KEEPALIVE メッセージを受信できなかったとき、コネクションがダウンしたものと判断する。
metric パラメータはすべての  MED の初期値として働くので、 bgp import  コマンドで  MED を設定したときにはそれ
が優先される。
gateway では、接続先が同一のセグメントにないときに、その接続先に対するゲートウェイ  ( ネクストホップ  ) を指
定する。
本コマンドは最大で 32個まで設定することができる。
vRX シリーズ  ではキープアライブを有効にすることで近隣ルーターおよび経路情報の更新が行われるため、 hold-
time パラメーターは  'off' 以外に設定する必要がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830474 | コマンドリファレンス  | BGP

29.13 BGP で使用する  TCP MD5 認証の事前共有鍵の設定
[書式 ]
bgp neighbor  pre-shared-key  neighbor_id  text text_key
no bgp neighbor  pre-shared-key  neighbor_id  [text text_key ]
[設定値及び初期値 ]
•neighbor_id
• [設定値 ] : 近隣ルーターの番号  (1...2147483647)
• [初期値 ] : -
•text_key
• [設定値 ] : ASCII 文字列で表した鍵  (80文字以内 )
• [初期値 ] : -
[説明 ]
BGP で使用する  TCP MD5 認証の事前共有鍵を設定する。設定した事前共有鍵が一致するピア間のみ、 BGPのコネ
クションが成立する。
[ノート ]
RTX1210 は Rev.14.01.11 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.14 BGP のログの設定
[書式 ]
bgp log log [log]
no bgp log [log ...]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
neighbor 近隣ルーターに対する状態遷移
packet 送受信したパケット
• [初期値 ] : ログを記録しない。
[説明 ]
指定した種類のログを  INFO レベルで記録する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.15 BGP で強制的に経路を広告する
[書式 ]
bgp force-to-advertise  remote_as  ip_address /mask  [parameter  ...]
no bgp force-to-advertise  remote_as  ip_address /mask  [parameter  ... ]
[設定値及び初期値 ]
•remote_as
• [設定値 ] : 相手の  AS 番号
• [初期値 ] : -
•ip_address/mask
• [設定値 ] : IP アドレス /ネットマスク
• [初期値 ] : -コマンドリファレンス  | BGP | 475

•parameter
• [設定値 ] :
• TYPE=V ALUE の組
TYPE V ALUE 説明
metric 1 .. 16777215MED (Multi-Exit Discriminator) で
通知するメトリック値
preference 0 .. 255同じ経路を複数の相手から受信し
たときに、一方を選択するための
優先度
• [初期値 ] : preference=100
[説明 ]
本コマンドで設定した経路がルーティングテーブルに存在しない場合でも、 指定された  AS 番号のルーターに対して
BGP で経路を強制的に広告する。経路として  'default' を指定した場合にはデフォルト経路が広告される。設定した
コマンドは  bgp configure refresh コマンドを実行したときに有効になる。
[ノート ]
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
RTX1210 は Rev.14.01.11 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
29.16 インタフェースの状態変化時、 BGP に外部経路を反映させる時間間隔の設定
[書式 ]
bgp reric  interval  time
no bgp reric  interval  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (1 以上の数値  )
• [初期値 ] : 1
[説明 ]
ルーターのインタフェースの状態が変化したとき、 BGP に外部経路を反映させる時間の間隔を設定する。
BGP ではインタフェースの状態変化を 1秒間隔で監視し、変化があれば最新の外部経路を自身に反映させるが、イ
ンタフェースの状態変化が連続して発生するときは、複数の外部経路の反映処理が  time で指定した秒数の間隔でま
とめて行われるようになる。
[ノート ]
複数のトンネルが一斉にアップすることがあるような環境では、本コマンドの値を適切に設定することで、 OSPF や
BGP の外部経路の導入によるシステムへの負荷を軽減することができる。
本コマンドの設定値は、 OSPF への外部経路の反映にも影響する。本コマンドと  ospf reric interval  コマンドの設定値
が食い違う場合には、 ospf reric interval  コマンドの設定値が優先して適用される。
本コマンドの設定は、 経路の変化や  IP アドレスの変化に対する  OSPF や BGP の動作には関係しない。また本コマン
ドの設定値は、 bgp configure refresh  コマンドを実行しなくても即時反映される。
RTX1210 は Rev.14.01.16 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.21 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830476 | コマンドリファレンス  | BGP

29.17 BGP の最適経路選択における MED属性が付加されていない経路のデフォルトの MED
値の設定
[書式 ]
bgp default  med med
no bgp default  med [med]
[設定値及び初期値 ]
•med
• [設定値 ] : MED値 (1..2147483647)
• [初期値 ] : 2147483647
[説明 ]
BGPの最適経路選択で、 MED属性が付加されていない経路に対するデフォルトの MED値を設定する。
本コマンドが設定されていない場合、 MED属性が付加されていない経路は最大の MED値(2147483647) を持つこと
になり、優先度は最低となる。
本コマンドの設定は、 MED属性が付加されている経路には影響しない。
[ノート ]
RTX5000 、RTX3500 は Rev.14.00.32 以降で使用可能。
RTX1210 は Rev.14.01.36 以降で使用可能。
RTX830 は Rev.15.02.15 以降で使用可能。
[適用モデル ]
vRX VMware ESXi 版, vRX さくらのクラウド版 , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210,
RTX840, RTX830
29.18 BGP で受信した経路に対する経路選択の優先順位の規則を設定
[書式 ]
bgp export  route  selection  rule rule
no bgp export  route  selection  rule [rule]
[設定値及び初期値 ]
•rule
• [設定値 ] :
設定値 説明
ebgp-only EBGPで受信した同じ宛先の経路を比較対象とする
all 全ての BGPで受信した同じ宛先の経路を比較対象と
する
• [初期値 ] : ebgp-only
[説明 ]
BGP で同じ宛先の経路を複数の相手から受信した際、一方を選択するための優先度による経路選択の規則を設定す
る。
本コマンドの設定により  bgp export filter  コマンドの  preference  で比較する経路の種別が変更される。
rule に ebgp-only を設定した場合、 EBGP で受信した経路に対してのみ  preference  による比較が働く。
rule に all を設定した場合、全ての  BGP で受信した経路に対して  preference  による比較が働く。このため、 IBGP で
受信した経路に対しても、 preference  による比較が働く。これにより、 IBGP で受信した経路の優先度を  EBGP で受
信した経路よりも高くすることが可能になる。
IBGP のみを使用した構成では、本コマンドの設定によって選択される経路が変わることはない。コマンドリファレンス  | BGP | 477

同じ近隣ルーター番号をもつ近隣ルータから複数の同一宛先経路を受信する場合、本コマンドによって規則を変更
した経路選択の優先度に差を設けることはできない。この場合、 EBGP であれば  MED を、 IBGP であれば  Local
Preference を使用して経路の優先度を設定できる。
本コマンドに対応していないリビジョンでは、 rule が ebgp-only のときの動作をする。
[ノート ]
本コマンドで動作を設定可能な経路選択プロセスは以下の  URL を参照してください。
• https://www.rtpro.yamaha.co.jp/RT/docs/bgp/index.html#rt_select
EBGPとIBGP間で同じ宛先の経路を受信する環境下で特定の経路を優先するよう制御したい場合、本コマンドを設
定することで実現できる。
RTX5000 / RTX3500 Rev.14.00.33 以前、 RTX1300 Rev.23.00.04 以前、 RTX1220 Rev.15.04.04 以前、 RTX1210 Rev.14.01.41
以前、および  RTX830 Rev.15.02.26 以前では使用不可能。
[適用モデル ]
vRX さくらのクラウド版 , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830478 | コマンドリファレンス  | BGP

