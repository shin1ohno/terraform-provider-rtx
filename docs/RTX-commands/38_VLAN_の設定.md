# 第38章: VLAN の設定

> 元PDFページ: 554-558

---

第 38 章
VLAN の設定
38.1 VLAN ID の設定
[書式 ]
vlan interface /sub_interface  802.1q vid= vid [name= name ]
no vlan interface /sub_interface  802.1q
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•sub_interface
• [設定値 ] : 1-32
• [初期値 ] : -
•vid
• [設定値 ] : VLAN ID(IEEE802.1Q タグの  VID フィールド格納値  ) (2 -4094)
• [初期値 ] : -
•name
• [設定値 ] : VLAN に付ける任意の名前  (半角で  127 文字以内、全角で  63 文字以内 )
• [初期値 ] : -
[説明 ]
LAN インタフェースで使用する  VLAN の VLAN ID を設定する。
設定された  VID を格納した  IEEE802.1Q タグ付きパケットを扱うことができる。
ひとつの  LAN インタフェースに設定できる  VLAN の数は  32個。
[ノート ]
タグ付きパケットを受信した場合、 そのタグの  VID が受信  LAN インタフェースに設定されていなければパケットを
破棄する。同一  LAN インタフェースで  LAN 分割機能  (lan type  コマンドの  port-based-option オプション ) との併用
はできない。両者のうち先に入力されたものが有効となり、後から入力されるものはコマンドエラーになる。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
38.2 スイッチングハブのポートが所属する  VLAN の設定
[書式 ]
vlan port mapping  sw_port  vlan_interface
no vlan port mapping  sw_port  [vlan_interface ]
[設定値及び初期値 ]
•sw_port
• [設定値 ] : スイッチングハブのポート  (lan1.1 - lan1.N 、lan2.1 - lan2.N)
• [初期値 ] : -
•vlan_interface
• [設定値 ] : VLAN インタフェース名  (vlan1 - vlanN)
• [初期値 ] : -
[説明 ]
LAN 分割機能の拡張機能において、スイッチングハブの各ポートが所属する  VLAN インタフェースを指定する。554 | コマンドリファレンス  | VLAN の設定

ポートの名称には  lan1.N / lan2.N を使用する。 lan2.N はスイッチインタフェースが  2 個ある機種で指定可能である。
同一の  VLAN インタフェースに所属するポート間はスイッチとして動作する。
RTX5000 、RTX3510 、RTX3500 では、 lan1.N のポートに対して  vlan1 ～ vlan4 をマッピングすることができ、 lan2.N
のポートに対しては  vlan5 ～ vlan8 をマッピングすることができる。
スイッチインタフェースが  1 個の機種の初期状態のマッピングは、 lan1.N ＝ vlanN となる。
RTX5000 、RTX3510 、RTX3500 の初期状態のマッピングを下表に示す。
ポート VLAN
LAN1.1 vlan1
LAN1.2 vlan2
LAN1.3 vlan3
LAN1.4 vlan4
LAN2.1 vlan5
LAN2.2 vlan6
LAN2.3 vlan7
LAN2.4 vlan8
[ノート ]
lan type  コマンドで  "port-based-option=divide-network" を設定し、 LAN 分割機能を有効にしなければ本コマンドは機
能しない。
"port-based-option=divide-network" の設定が無い場合でも vlan port mapping  は設定できるが、スイッチングハブの動
作は変化しない。
[設定例 ]
# vlan port mapping lan1.3 vlan2
# vlan port mapping lan1.4 vlan2
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1220, RTX1210, RTX840, RTX830
38.3 VLAN 相互接続インタフェースグループの設定
[書式 ]
vlan interconnect  group  phys_lan_interface  group  [group ...]
no vlan interconnect  group  phys_lan_interface  [group ...]
[設定値及び初期値 ]
•phys_lan_interface  : グループの設定対象となる物理  LAN インタフェース
• [設定値 ] : 物理  LAN インタフェース名
• [初期値 ] : -
•group  : グループを構成する  LAN インタフェース群
• [設定値 ] :
• all
•phys_lan_interface  に従属するすべての  LAN インタフェースの相互接続を許可する（全開放）
• none
•phys_lan_interface  に従属するすべての  LAN インタフェースの相互接続を禁止する（全遮断）
•物理  LAN インタフェース名、および、仮想  LAN インタフェース名を「 ,」で連結して指定
•列挙した  LAN インタフェースの相互接続を許可する
•仮想  LAN インタフェース名を「 -」 （ハイフン）で連結して指定
•連続した仮想  LAN インタフェースの相互接続を許可する
•物理  LAN インタフェース名、または、仮想  LAN インタフェース名に「 $」 （ドルマーク）を付与して指定コマンドリファレンス  | VLAN の設定  | 555

•指定した  LAN インタフェースのみに  phys_lan_interface  に従属する他のすべての  LAN インタフェース
との相互接続を許可する
• [初期値 ] : すべての物理  LAN インタフェースで  all
[説明 ]
物理  LAN インタフェースごとに、相互に接続可能な  LAN インタフェースのグループを設定する。
仮想  LAN インタフェース（  LAN 分割インタフェース、または、タグ  VLAN インタフェース）を使用する場合、初
期状態では物理  LAN インタフェースとの間の通信、および、すべての仮想  LAN インタフェースとの間の通信が可
能であるが、本コマンドによって相互に通信を許可する  LAN インタフェースのグループを任意に指定できる。
group  に all を指定した場合は、 phys_lan_interface  に従属するすべての  LAN インタフェースの相互接続が可能となる
（全開放） 。 group に none を指定した場合は、 phys_lan_interface  に従属するすべての  LAN インタフェースが互いに遮
断され、相互接続ができなくなる（全遮断） 。ただし、 none を指定した場合でも、 phys_lan_interface に指定した物理
LAN インタフェースには従属しない他の物理  LAN インタフェース、および、仮想  LAN インタフェースへの通信は
遮断されない。なお、 all / none の指定は本コマンド実行後に作成した仮想  LAN インタフェースに対しても有効にな
るため、仮想  LAN インタフェースを増やす度に本コマンドを実行する必要はない。
相互接続を許可するグループを任意に設定する場合は、グループごとに複数の物理  LAN インタフェース名、およ
び、仮想  LAN インタフェース名を「 -」 （ハイフン）または「 ,」 （カンマ）で連結した  LAN インタフェース群を  group
に指定する。
「$」 （ドルマーク）を使用すれば、すべての  LAN インタフェースとのペアを簡略化して表現することができる。例
えば、 LAN 分割インタフェース  6 個（ VLAN1～VLAN6）を使用しているとき、 VLAN2 インタフェースのみが他の
すべての  LAN 分割インタフェースとの相互接続を可能とする場合、 group は「 vlan1,vlan2 vlan2,vlan3 vlan2,vlan4
vlan2,vlan5 vlan2,vlan6 」のように  5 個のグループを指定する必要があるが、これを「 vlan2$」と簡略化して指定する
ことができる。 「 $」はすべての  LAN インタフェースと個別にペアになった複数のグループへ展開されることを意味
する。同様に、タグ  VLAN インタフェース  10 個（ LAN1/1～LAN1/10 ）を使用しているとき、 lan1/1 インタフェース
のみが他のすべてのタグ  VLAN インタフェースおよび  Native VLAN インタフェース（タグ  VLAN を使用しない
LAN1 インタ ―フェース） との相互接続を可能とする場合、 group は 「lan1,lan1/1 lan1/1,lan1/2 lan1/1,lan1/3 lan1/1,lan1/4
lan1/1,lan1/5 lan1/1,lan1/6 lan1/1,lan1/7 lan1/1,lan1/8 lan1/1,lan1/9 lan1/1,lan1/10 」 のように  10 個のグループを指定する必
要があるが、これを「 lan1/1$」と簡略化して指定することができる。 「 $」も  all / none の指定と同様に、本コマンド
実行後に作成した仮想  LAN インタフェースに対しても有効になるため、仮想  LAN インタフェースを増やす度に本
コマンドを実行する必要はない。
本コマンドに似た機能として  LAN 分割拡張機能のポート分離機能があるが、ポート分離機能は同一の仮想  LAN イ
ンタフェース内の通信（内蔵スイッチングハブ内で折り返される通信）を制御する機能であり、複数の仮想  LAN イ
ンタフェースをまたぐ通信は制御できない。本コマンドは複数の仮想  LAN インタフェースをまたぐ通信を制御で
きる。
LAN インタフェースと  LAN インタフェース以外のインタフェース（  PP インタフェースや  TUNNEL インタフェー
スなど）との間の通信、および、複数の物理  LAN インタフェースをまたぐ通信は本コマンドによる制御の対象外で
ある。
IP フィルター機能と併用する場合は、 OUT 側 IP フィルターのチェック処理の後で本コマンドによる通信制御処理が
行われ、 IP フィルター機能と本コマンドの両方で通信が許可されているパケットがインタフェース間を通過でき
る。
[ノート ]
RTX3510 は Rev.23.01.02 以降で使用可能。
RTX1300 は Rev.23.00.12 以降で使用可能。
RTX1220 は Rev.15.04.07 以降で使用可能。
RTX830 は Rev.15.02.31 以降で使用可能。556 | コマンドリファレンス  | VLAN の設定

[設定例 ]
• LAN1 インタフェースを全開放
# vlan interconnect group lan1 all
• LAN2 インタフェースを全遮断
# vlan interconnect group lan2 none
• vlan1、vlan2、vlan3 のグループと  vlan2、vlan3、vlan4 のグループで各グループ内の相互接続を許可
# vlan interconnect group lan1 vlan1-3 vlan2-4
• vlan2 のみで  vlan1 および  vlan3～vlan8 との相互接続を許可
# vlan interconnect group lan1 vlan2$
• lan2/1 、lan2/2、lan2/3、lan2/18 で相互接続を許可
# vlan interconnect group lan2 lan2/1-3,lan2/18
• lan2 と lan2/1 のみですべての  LAN2 インタフェースに従属する  LAN インタフェースとの相互接続を許可
# vlan interconnect group lan2 lan2$ lan2/1$
[適用モデル ]
RTX3510, RTX1300, RTX1220, RTX840, RTX830
38.4 VLAN 相互接続インタフェースグループによる通信制御のログを記録するか否か
[書式 ]
vlan interconnect  log phys_lan_interface  type [type]
no vlan interconnect  log phys_lan_interface  [type [type]]
[設定値及び初期値 ]
•phys_lan_interface  : グループの設定対象となる物理  LAN インタフェース
• [設定値 ] : 物理  LAN インタフェース名
• [初期値 ] : -
•type
• [設定値 ] :
設定値 説明
pass 相互接続が許可されたインタフェース間の通信に関
するログを記録する
reject 相互接続が許可されていないインタフェース間の遮
断された通信に関するログを記録する
• [初期値 ] : -
[説明 ]
VLAN 相互接続インタフェースグループによる通信制御のログを記録するか否かを設定する。
本コマンドを設定した場合、 VLAN 相互接続インタフェースグループで通信制御が行われた際に  NOTICE レベルの
syslog を出力する。
[ノート ]
IP フィルター機能と  VLAN 相互接続インターフェースグループによる通信制御を併用する場合は、 先に  IP フィルタ
ーのチェック処理が行われる。そのため、 IP フィルターでパケットが破棄された場合は、本コマンドによるログは
出力されない。
RTX3510 は Rev.23.01.02 以降で使用可能。
RTX1300 は Rev.23.00.12 以降で使用可能。
RTX1220 は Rev.15.04.07 以降で使用可能。
RTX830 は Rev.15.02.31 以降で使用可能。コマンドリファレンス  | VLAN の設定  | 557

[適用モデル ]
RTX3510, RTX1300, RTX1220, RTX840, RTX830558 | コマンドリファレンス  | VLAN の設定

