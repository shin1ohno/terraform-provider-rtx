# 第52章: DPI

> 元PDFページ: 694-707

---

第 52 章
DPI
Deep Packet Inspection ( 以下  DPI) は、 IP ネットワーク上を流れるトラフィックを高度に検査することにより、そのパ
ケットがどのアプリケーションのものであるかを識別することができる技術です。
ヤマハルーターでは、 DPI の識別結果を以下のように利用することができます。
•経路の選択
•フィルタリング
• QoS
•トラフィックの可視化
DPI は有償サービスにより提供される機能です。ご利用いただくには、別途ライセンス製品を購入していただく必
要があります。
本機能に関する技術情報は以下に示す  URL で公開しています。
https://www.rtpro.yamaha.co.jp/RT/docs/dpi/
52.1 DPI を使用するか否かの設定
[書式 ]
dpi use switch  [reject ]
no dpi use [switch  [reject ]]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
•reject  : 未識別のパケットを破棄するか否か
• [設定値 ] :
設定値 説明
on 破棄する
off 破棄しない
• [初期値 ] : off
[説明 ]
DPI を使用するか否かを設定する。
DPI のアクティベーション中でアプリケーションの識別ができない場合のパケットは、  reject パラメーターの設定に
従い通過、または破棄される。
[ノート ]
RTX830 の Rev.15.02.26 以前、および、  RTX1300 の Rev.23.00.04 以前  のファームウェアで、アプリケーション識別
を行うためには、すべてのパケットをノーマルパスで処理する必要がある。 IPv4 の環境では、 ip routeing process  コ
マンドを、 IPv6 の環境では  ipv6 routing process  コマンドを  normal に設定する必要がある。
RTX830 は Rev.15.02.13 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830694 | コマンドリファレンス  | DPI

52.2 IPv4 の DPI フィルターの設定
[書式 ]
ip dpi filter  filter_num  pass_reject  src_addr [/mask ] [dest_addr [/mask ] [application ]]
ip dpi filter  filter_num  pass_reject  src_addr [/mask ] [dest_addr [/mask ] [group_num ]]
no ip dpi filter  filter_num  [pass_reject  ...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : DPI のフィルター番号  (1..2147483647)
• [初期値 ] : -
•pass_reject
• [設定値 ] :
設定値 説明
pass 一致すれば通す (ログに記録しない )
pass-log 一致すれば通す (ログに記録する )
pass-nolog 一致すれば通す (ログに記録しない )
reject 一致すれば破棄する (ログに記録しない )
reject-log 一致すれば破棄する (ログに記録する )
reject-nolog 一致すれば破棄する (ログに記録しない )
• [初期値 ] : -
•src_addr  : IP パケットの始点アドレス
• [設定値 ] :
• IPv4 アドレス
• A.B.C.D (A ～ D: 0 ～ 255 もしくは  *)
•上記表記で  A ～ D を * とすると、該当する  8 ビット分についてはすべての値に対応する
•間に  - を挟んだ  2 つの上項目、 - を前につけた上項目、 - を後ろにつけた上項目、これらは範囲を指定す
る
•カンマ区切りで複数設定することができる
• FQDN
•任意の文字列  (半角  255 文字以内。  /,: は使用できない。  , は区切り文字として使われるため、使用でき
ない )
• * から始まる  FQDN は * より後ろの文字列を後方一致条件として判断する。たとえば  *.example.co.jp
は www.example.co.jp 、mail.example.co.jp などと一致する
• , を区切りとして複数設定することができる。 IP アドレスと混在することも可能
• * (すべてのアドレスに対応 )
• [初期値 ] : -
•dest_addr  : IP パケットの終点アドレス
• [設定値 ] :
•src_addr  と同じ形式
•省略した場合は  1 個の  * と同じ
• [初期値 ] : -
•mask  : ネットワークアドレスのビットマスク
• [設定値 ] :
• A.B.C.D (A ～ D: 0 ～ 255)
• 0x に続く十六進数
•マスクビット数
•省略時は  0xffffffff と同じ
• [初期値 ] : -
•application  : フィルタリング対象とするアプリケーション、またはカテゴリー
• [設定値 ] :コマンドリファレンス  | DPI | 695

•アプリケーションを表すニーモニック
• "@" で始まるカテゴリーをあらわすニーモニック
•上記文字列をカンマで区切った並び  (10 個以内、アプリケーションとカテゴリーの混在が可能 )
•省略時は 1つの  * と同じ
• [初期値 ] : -
•group_num  : グループ  ID
• [設定値 ] : dpi group set  コマンドでグループ化したアプリケーションのグループ  ID
• [初期値 ] : -
[説明 ]
DPI で使用する  IPv4 のフィルターを設定する。本コマンドで設定されたフィルターは  ip interface dpi filter  コマン
ドと組み合わせて利用することで、特定アプリケーションのパケットのフィルタリングをすることができる。また、
フィルター型ルーティングと利用することで、特定アプリケーションのパケットの経路選択を行うことができる。
アプリケーションを表すニーモニックには、 show dpi application  コマンドで表示されるものを使用する。またカテ
ゴリーを表すニーモニックには、 show dpi category  コマンドで表示されるものを使用する。
[ノート ]
アプリケーションの識別が完了していないパケットは、フィルターの設定によらず必ず通過する。また、 DPI のア
クティベーション中でアプリケーションの識別結果が得られない場合には、すべてのパケットは  dpi use  コマンドの
reject  の設定に従う。
RTX830 は Rev.15.02.13 以降で使用可能。
[設定例 ]
• Web アクセスとメールの使用のみ許可する
# ip dpi filter 1 pass * * @web,@mail,@webmail
# ip dpi filter 100 reject * *
# ip lan2 dpi filter out 1 100
• Office365 のパケットは  PP1 インターフェース経由で、その他は  TUNNEL1 経由で送信する
# dpi group set 100 name=o365 word_online sharepoint_online powerpoint_online 
outlook office_docs office365 ms_sway ms_planner ms_onenote lync_online 
excel_online
# ip dpi filter 1 pass * * 100
# ip route default gateway pp 1 dpi 1 gateway tunnel 1
[適用モデル ]
RTX1300, RTX840, RTX830
52.3 IPv6 の DPI フィルタの設定
[書式 ]
ipv6 dpi filter  filter_num  pass_reject  src_addr [/prefix_len ] [dest_addr [/prefix_len ][application ]]
ipv6 dpi filter  filter_num  pass_reject  src_addr [/prefix_len ] [dest_addr [/prefix_len ][group_num ]]
no ipv6 dpi filter  filter_num  [pass_reject  ...]
[設定値及び初期値 ]
•filter_num
• [設定値 ] : DPI のフィルタ番号  (1..2147483647)
• [初期値 ] : -
•pass_reject
• [設定値 ] :
設定値 説明
pass 一致すれば通す (ログに記録しない )
pass-log 一致すれば通す (ログに記録する )
pass-nolog 一致すれば通す (ログに記録しない )
reject 一致すれば破棄する (ログに記録しない )696 | コマンドリファレンス  | DPI

設定値 説明
reject-log 一致すれば破棄する (ログに記録する )
reject-nolog 一致すれば破棄する (ログに記録しない )
• [初期値 ] : -
•src_addr  : IPv6 パケットの始点アドレス
• [設定値 ] :
• IPv6 アドレス
•間に  - を挟む、 - を前につける、または  - を後ろにつける範囲の指定ができる
•カンマ区切りで複数設定することができる
• *(すべてのアドレスに対応 )
• [初期値 ] : -
•dest_addr  : IPv6 パケットの終点アドレス
• [設定値 ] :
•src_addr  と同じ形式
•省略した場合は  1 個の  * と同じ
• [初期値 ] : -
•prefix_len
• [設定値 ] : プレフィックス長
• [初期値 ] : -
•application  : フィルタリング対象とするアプリケーション、またはカテゴリー
• [設定値 ] :
•アプリケーションを表すニーモニック
• "@" で始まるカテゴリーをあらわすニーモニック
•上記文字列をカンマで区切った並び  (10 個以内、アプリケーションとカテゴリーの混在が可能 )
•省略時は 1つの  * と同じ
• [初期値 ] : -
•group_num  : グループ  ID
• [設定値 ] : dpi group set  コマンドでグループ化したアプリケーションのグループ  ID
• [初期値 ] : -
[説明 ]
DPI で使用する  IPv6 フィルタを設定する。本コマンドで設定されたフィルタは  ipv6 interface dpi filter  コマンドと
組み合わせて利用することで、特定アプリケーションのパケットのフィルタリングをすることができる。また、フ
ィルタ型ルーティングと利用することで、特定アプリケーションのパケットの経路選択を行うことができる。
アプリケーションを表すニーモニックには、 show dpi application  コマンドで表示されるものを使用する。またカテ
ゴリーを表すニーモニックには、 show dpi category コマンドで表示されるものを使用する。
[ノート ]
アプリケーションの識別が完了していないパケットは、フィルタの設定によらず必ず通過する。また、 DPI のアク
ティベーション中でアプリケーションの識別結果が得られない場合には、すべてのパケットは  dpi use  コマンドの
reject  の設定に従う。
RTX830 は Rev.15.02.13 以降で使用可能。
[設定例 ]
• IPv6 による  FTP サーバーへのアクセスは禁止する
# ipv6 dpi filter 1 reject * * ftp,ftp_data
# ipv6 dpi filter 100 pass * *
# ipv6 lan2 dpi filter out 1 100
• Office365 のパケットは  DHCP クライアントとして動作する  LAN2 インターフェース経由で、 その他は  TUNNEL1
経由で送信する
# dpi group set 100 name=o365 word_online sharepoint_online powerpoint_online 
outlook office_docs office365 ms_sway ms_planner ms_onenote lync_online コマンドリファレンス  | DPI | 697

excel_online
# ipv6 dpi filter 1 pass * * 100
# ipv6 route default gateway dhcp lan2 dpi 1 gateway tunnel 1
[適用モデル ]
RTX1300, RTX840, RTX830
52.4 DPI のフィルターのインターフェースへの適用
[書式 ]
ip interface  dpi filter  direction  filter_list ...
ipv6 interface  dpi filter  direction  filter_list ...
ip pp dpi filter  direction  filter_list ...
ipv6 pp dpi filter  direction  filter_list ...
ip tunnel  dpi filter  direction  filter_list ...
ipv6 tunnel  dpi filter  direction  filter_list ...
no ip interface  dpi filter  direction  [filter_list ...]
no ipv6 interface  dpi filter  direction  [filter_list ...]
no ip pp dpi filter  direction  [filter_list ...]
no ipv6 pp dpi filter  direction  [filter_list ...]
no ip tunnel  dpi filter  direction  [filter_list ...]
no ipv6 tunnel  dpi filter  direction  [filter_list ...]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インターフェース名、 LOOPBACK インターフェース名、 NULL イン
ターフェース名、ブリッジインターフェース名
• [初期値 ] : -
•direction
• [設定値 ] :
設定値 説明
in 受信したパケットに対するフィルタリング
out 送信したパケットに対するフィルタリング
• [初期値 ] : -
•filter_list
• [設定値 ] : 空白で区切られた静的フィルタ番号の並び  (最大 128 個以内  )
• [初期値 ] : -
[説明 ]
ip dpi filter  コマンド、および  ipv6 dpi filter  コマンドによるフィルターを組み合わせて、インターフェースで送受信
するパケットの種類を制限する。
送信 /受信のそれぞれの方向に対して、適用するフィルター列をフィルター番号で指定する。指定された番号のフィ
ルターが順番に適用され、パケットにマッチするフィルターが見つかればそのフィルターにより通過 /破棄が決定す
る。それ以降のフィルターは調べられない。すべてのフィルターにマッチしないパケットは破棄される。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[設定例 ]
• 192.168.200.0/24 のネットワークに属する端末で、指定のファイル共有ソフトと、ゲームカテゴリーに属するアプ
リケーションを禁止する
# dpi group set 1000 name=file-sharing winmx winny share bittorrent
# ip dpi filter 1 reject 192.168.200.0/24 * 1000
# ip dpi filter 2 reject 192.168.200.0/24 * @game
# ip dpi filter 100 pass * *
# pp select 1
#  ip pp dpi filter out 1 2 100698 | コマンドリファレンス  | DPI

[適用モデル ]
RTX1300, RTX840, RTX830
52.5 DPI のアプリケーショングループの作成
[書式 ]
dpi group  set group_num  [name= name ] application_list ...
no dpi group  set group_num  [[name= name ] application_list ...]
[設定値及び初期値 ]
•group_num
• [設定値 ] : グループ  ID (1..2147483647)
• [初期値 ] : -
•name
• [設定値 ] : グループ名  (最大  32 文字以内 )
• [初期値 ] : -
•application_list
• [設定値 ] : アプリケーション、およびカテゴリーの並び (空白区切り、最大  128 個以内 )
• [初期値 ] : -
[説明 ]
DPI のアプリケーションのグループを作成する。本コマンドで作成したグループは、 以下の各コマンドで  group キー
ワードに続いて指定することができる。
•ip dpi filter
•ipv6 dpi filter
•queue class filter
application_list  には、アプリケーションやカテゴリーを表すニーモニックを並べて指定する。
name  は、半角英数字、 "-"(ハイフン )、および "_"(アンダースコア )で、最大  32 文字以内で指定する。グループはルー
ターの揮発性メモリが許す限り作成することができる。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[設定例 ]
•音楽ストリーミングアプリケーションを禁止する
# dpi group set 1000 name=music-streaming spotify google_play_music apple_music 
amazon_music
# ip dpi filter 1 reject 1000
# ip dpi filter 100 pass * *
# ip lan2 dpi filter out 1 100
[適用モデル ]
RTX1300, RTX840, RTX830
52.6 シグネチャーのダウンロードの手動実行
[書式 ]
dpi signature  download  go [no-confirm [prompt]] [force]
[設定値及び初期値 ]
• no-confirm : 更新可能なシグネチャーが存在するときに、シグネチャーの更新を行うか否かを確認しない
• [初期値 ] : -
• prompt : コマンド実行後、すぐにプロンプトを表示させ、他のコマンドを実行できるようにする
• [初期値 ] : -
• force : 新しいシグネチャーの有無のチェックを行わず、強制的にシグネチャーをダウンロードする
• [初期値 ] : -コマンドリファレンス  | DPI | 699

[説明 ]
DPI のシグネチャーのダウンロードや更新の手動実行をする。
シグネチャーのダウンロードに一度も成功していない状態で本コマンドを実行すると、配布サーバーからシグネチ
ャーのダウンロードを試みる。ダウンロードに成功した場合、シグネーチャーが  DPI エンジンにロードされて  DPI
が使用可能な状態になる。
シグネチャーがダウンロードされている状態で本コマンドを実行した場合、配布サーバーに対して新しいシグネチ
ャーの有無のチェックをして、更新可能なシグネチャーが存在すれば「更新しますか？ (Y/N)」の確認を求める。 "Y"
を入力するとダウンロード、およびロードを行う。 "N" を入力すると、更新を中止する。更新可能なシグネチャーが
なければ、 「新しいシグネチャーはありません。 」と表示する。
no-confirm を指定すると、更新可能なシグネチャーが存在する場合に更新を行うか否かの確認を行わない。 prompt
を指定すると、コマンド実行直後にプロンプトが表示され、続けて他のコマンドを実行することができるようにな
る。
force を指定した場合には、新しいシグネチャーの有無のチェックを行わず、強制的にシグネチャーをダウンロード
する。外部メモリにシグネチャーを保存したい場合に有効である。
新しいシグネチャーの有無のチェックやダウンロードに失敗した場合でも、リトライはしない。
本コマンドは、 dpi use  コマンドが  off に設定されている場合には実行できない。
[ノート ]
シグネチャーのダウンロードや更新は自動で行われるため、通常は本コマンドを実行する必要はない。自動でのダ
ウンロードに失敗したときなどに、直ちにリトライした場合に使用する。
RTX830 は Rev.15.02.13 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830
52.7 シグネチャーのダウンロード先  URL の設定
[書式 ]
dpi signature  download  url url
no dpi signature  download  url [url]
[設定値及び初期値 ]
•url
• [設定値 ] : シグネチャーを配布している  URL を設定する
• [初期値 ] :
• http://www.rtpro.yamaha.co.jp/signature/v1.1/rt_dpi_( 機種名 ).ysig ( RTX830 Rev.15.02.29 以降  )
• http://www.rtpro.yamaha.co.jp/signature/rt_dpi_( 機種名 ).ysig ( 上記以外  )
[説明 ]
個別に用意した  Web サーバーを用いてシグネチャーを配布する場合に、 シグネチャーが置かれている  URL を 255 文
字以内の半角英数字および半角記号で指定する。ヤマハの配信サーバーからシグネチャーをダウンロードして使用
する場合には、本コマンドを設定する必要はない。
シグネチャーのダウンロードには、 HTTP または  HTTPS を使用できる。入力形式は以下の通りで、 Web サーバーの
アドレスは  FQDN 形式のホスト名、もしくは  IPv4 または  IPv6 アドレスを指定する。
• http[s]://(Web サーバーのアドレス )[:(ポート番号 )]/(パス名 )
ポート番号は  HTTP の場合  80 番以外、 HTTPS の場合  433 番以外を使用するときに指定する。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830
52.8 アプリケーションの識別に関するログを出力するか否かの設定
[書式 ]
dpi log switch700 | コマンドリファレンス  | DPI

no dpi log [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 出力する
off 出力しない
• [初期値 ] : off
[説明 ]
DPI のアプリケーションの識別に関するログを出力するか否かの設定をする。
本コマンドを  on に設定すると、 アプリケーションの識別結果や識別処理に関するエラー情報を  NOTICE レベルのロ
グに出力する。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830
52.9 DPI の統計情報の表示
[書式 ]
show  dpi statistics  [all]
show  dpi statistics  application application  [detail]
show  dpi statistics  category category
show  dpi statistics  address ip_address
[設定値及び初期値 ]
•application
• [設定値 ] : アプリケーションを表すニーモニック
• [初期値 ] : -
•category
• [設定値 ] : "@"で始まるカテゴリーを表すニーモニック
• [初期値 ] : -
•ip_address
• [設定値 ] : 端末の  IPv4 または  IPv6 アドレス
• [初期値 ] : -
[説明 ]
DPI の統計情報を表示する。
show dpi statistics  コマンドをオプション無しで実行した場合、ルーターを通過するトラフィック全体に対して送信
オクテット数と受信オクテットするの合計が上位 11位以内のアプリケーションの送受信オクテット数と、全体に占
める割合の高いものから表示する。 all を指定した場合には、 11 位以下の "other"にまとめられたアプリケーションの
一覧を表示する
application に続いて  application  を指定した場合には、指定のアプリケーションに関する統計情報のみを表示する。
application  には  show dpi application  コマンドで表示されるアプリケーションのニーモニックを指定する。 detail を
指定した場合、端末ごとの情報も表示する。
category に続いて  category  を指定した場合には、 指定のカテゴリーに属するアプリケーションに関する統計情報のみ
を表示する。 category  には "@" で始まる  show dpi category  コマンドで表示されるカテゴリーのニーモニックを指定
する。  カテゴリー内のアプリケーションは、アプリケーションのニーモニックを基準に、 0-9、a-z の準に表示する。
address に続いて ip_address  に端末の  IPv4 アドレスまたは  IPv6 アドレスを指定した場合には、指定した端末の統計
情報を表示する。コマンドリファレンス  | DPI | 701

[ノート ]
統計情報を記録できる端末の台数には上限がある。機種別の上限の台数は以下の通り。
機種 上限の台数  (台)
RTX1300 512
RTX840 512
RTX830 512
ルーターの揮発性メモリに記録されている統計情報は、以下の条件でクリアされる。
•ルーターの電源断
•ルーターの再起動
•dpi use  コマンドを  off に設定する、または削除する
•clear dpi statistics  コマンドの実行
RTX830 は Rev.15.02.13 以降で使用可能。
[表示例 ]
# show dpi statistics
アプリケーション   送受信オクテット数     割合     
----------------+---------------------+-----------
office365                  18,754,987    25%
lync                       15,854,347    21%
google_map                 11,115,753    14%
facebook                    9,835,354    11%
windows_update              6,354,682     8%
twitter                     5,854,965     6%
dns                         5,032,886     5%
ftp                         4,258,369     4%
icmp                        3,125,784     3%
imap                        1,411,955     1%
other                       1,923,854     2%
----------------+---------------------+-----------
合計                      　83,522,936   100%
# show dpi statistics application office365 detail
[office365]
                         　　オクテット数
端末                     送信           受信
-------------------+-------------+-------------
  Total                    88,258      131,564
   192.168.0.15            11,254       10,223
   192.168.0.68            48,256       58,825
            :
            :
# show dpi statistics category @game
[@game]
                             オクテット数
アプリケーション           送信           受信
-------------------+-------------+-------------
akinator                      456        1,675
all_slots_casino           11,254       10,223
            :
            :
# show dpi statistics address 192.168.0.15
[192.168.0.15]
                             オクテット数
アプリケーション           送信           受信
-------------------+-------------+-------------
lync                       32,645    1,556,998
office365                     546      665,320702 | コマンドリファレンス  | DPI

            :
            :
[適用モデル ]
RTX1300, RTX840, RTX830
52.10 DPI の統計情報のクリア
[書式 ]
clear  dpi statistics
[説明 ]
ルーターの揮発性メモリに保存された  DPI の統計情報をクリアする。
本コマンドを実行しても、外部メモリに記録された統計情報が削除されることはない。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830
52.11 識別結果のキャッシュの表示
[書式 ]
show  dpi cache
[説明 ]
アプリケーションの識別結果のキャッシュを IPv4／IPv6アドレスに分けて、 0-9、a-zの順に表示する。  TTLは分 ′秒″
のフォーマットで表示する。
[ノート ]
RTX830 は Rev.15.02.15 以降で使用可能。
[表示例 ]
# show dpi cache
エントリー数 : 6
アプリケーション名        宛先IPアドレス                            ポート  TTL
------------------------+----------------------------------------+------+------
microsoft                203.0.113.254                            443    58'41"
ntp                      203.0.113.110                            123    51'38"
office365                203.0.113.31                             443     6'25"
microsoft                2001:0db8:09ec:a541:20f7:3ba8:2808:bfd5  443    52'34"
ntp                      2001:0db8:e201:bc01:2951:c821:ba11:bbdd  123    12'56"
pokemon_go               2001:0db8:02c1:1bbb:5489:cde9:0001:ac19  443     2'12"
[適用モデル ]
RTX1300, RTX840, RTX830
52.12 識別結果のキャッシュのクリア
[書式 ]
clear  dpi cache
[説明 ]
アプリケーションの識別結果のキャッシュをクリアする。
[ノート ]
RTX830 は Rev.15.02.15 以降で使用可能。
[適用モデル ]
RTX1300, RTX840, RTX830コマンドリファレンス  | DPI | 703

52.13 アプリケーションを表すニーモニック一覧の表示
[書式 ]
show  dpi application  [category ]
show  dpi application  detail [ category ]
[設定値及び初期値 ]
•category
• [設定値 ] : "@"で始まるカテゴリーを表すニーモニック
• [初期値 ] : -
• detail : ニーモニックに加えて、アプリケーションの詳細情報を表示する
• [初期値 ] : -
[説明 ]
ip dpi filter  コマンド、 ipv6 dpi filter  コマンド、および、 queue class filter  コマンドで指定可能なアプリケーションを
表すニーモニック一覧を  0-9、a-z の順に表示する。
category  を指定した場合には、該当カテゴリーに属するアプリケーションを表すニーモニックのみを表示する。
category  には、 show dpi category  コマンドで表示される  "@" で始まるニーモニックを指定する。 detail  キーワードを
指定した場合には、アプリケーションの詳細情報を表示する。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[表示例 ]
# show dpi application
01net            050plus            0zz0            10050net
10086cn            104com            1111tw            114la
115com            118114cn        11st            123people
1337x            139mail            15min            163com
17173com        17u            20min            24h
24ora            24sata            24ur            2ch
2shared            3366_com        360buy            360cn
3gpp_li            3pc            4399com            4chan
4shared            4tube            51_com            51_com_bbs
51_com_music        51_com_posting        51job            51la
---つづく ---
# show dpi application detail @game
ニーモニック             説明
-----------------------+-------------------------------------------------------
akinator        Akinator the Genie
all_slots_casino    All Slots Casino
angry_birds        Angry Birds
aniping            Anipang
battlenet        battlenet
bf1            Battlefield 1
bf4            Battlefield 4
bitstrips        BitStrips
candy_crush_saga    Candy Crush Saga
champion_football    Champion Football
---つづく ---
[適用モデル ]
RTX1300, RTX840, RTX830
52.14 カテゴリーを表すニーモニック一覧の表示
[書式 ]
show  dpi category
show  dpi category  detail
[設定値及び初期値 ]
• detail : ニーモニックに加えて、カテゴリーの詳細情報を表示する704 | コマンドリファレンス  | DPI

• [初期値 ] : -
[説明 ]
ip dpi filter  コマンド、 ipv6 dpi filter  コマンド、および、 queue class filter  コマンドで指定可能なカテゴリーを表すニ
ーモニック一覧を  0-9、a-z の順に、先頭に  "@" を付加して表示する。
detail  キーワードを指定した場合には、ニーモニックに加えてカテゴリーの詳細情報を表示する。
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[表示例 ]
# show dpi category
@antivirus        @app_service        @audio_video        @authentication
@behavioral        @compression        @database        @encrypted
@erp            @file_server        @file_transfer        @forum
@game            @instant_messaging    @mail            @microsoft_office
@middleware        @network_management    @network_service    @peer_to_peer
@printer        @routing        @security_service    @standard
@telephony        @terminal        @thin_client        @tunneling
@wap            @web            @webmail
# show dpi category detail
ニーモニック     詳細
---------------+---------------------------------------------------------------
@antivirus    Antivirus update
@app_service    Background service
@audio_video    Application/Protocol used to transport audio or video content
@authentication    Protocol used for authentification purposes
@behavioral    Protocol classified by non-deterministic criteria based on 
statistical analysis of packet form and session behavior. 
@compression    Compression layers
@database    Protocol used for database remote queries
@encrypted    Encryption protocol
@erp        Enterprise Resource Planning application
@file_server    File transfer protocol
---つづく ---
[適用モデル ]
RTX1300, RTX840, RTX830
52.15 DPI の動作状態、およびシグネチャーの状態の表示
[書式 ]
show  status  dpi
[説明 ]
DPI の動作状態、およびシグネチャーの情報を表示する。
• DPIの状態
表示 説明
無効 DPIの設定は無効である
ライセンス認証中 ライセンス認証中である
シグネチャーのダウンロード中 シグネチャーをダウンロードしている
シグネチャーの読み込み中 シグネチャーを読み込んでいる
正常動作中 DPI は正常に動作している
停止処理中 DPI の停止処理をしている
ライセンス認証失敗ライセンス認証に失敗して停止している、ライセンス
認証の成功を待っているコマンドリファレンス  | DPI | 705

表示 説明
シグネチャーのダウンロード失敗  (リトライ待ち )シグネチャーのダウンロードに失敗し、リトライを待
っている
異常停止中  (理由 )エラーが発生し、 DPIは停止している (括弧内には異常
理由を表示 )
•シグネチャー情報
•バージョン：シグネチャーのバージョン情報
•ダウンロード日時：現在使用中のシグネチャーをダウンロードした日時
•外部メモリのシグネチャーを使用している場合には "外部メモリーのシグネチャーを使用中 "と表示する
•最終更新チェック日時：最後にシグネチャーの更新チェックを行った日時
•保存先：ダウンロードしたシグネチャーの外部メモリ上の保存先
•保存しない場合には  "-" を表示する
[ノート ]
RTX830 は Rev.15.02.13 以降で使用可能。
[表示例 ]
# show status dpi
現在の状態
    正常動作中
シグネチャーの情報
    バージョン             Ver. 1.0.6
    最終ダウンロード日時     外部メモリーのシグネチャーを使用中
    最終更新チェック日時     2019/10/13 09:14:25
    保存先             usb1:/signature/rt_dpi_rtx830.ysig
[適用モデル ]
RTX1300, RTX840, RTX830
52.16 シグネチャーを保存するディレクトリーの設定
[書式 ]
dpi signature  directory  path
no dpi signature  directory  [path]
[設定値及び初期値 ]
•path
• [設定値 ] :
設定値 説明
usb1:/directory_path USB メモリ内のディレクトリー
sd1:/directory_path microSD カード内のディレクトリー
/directory_path 内蔵不揮発性メモリー（ RTFS）内のディレクトリー
• [初期値 ] : -
[説明 ]
シグネチャーを保存するディレクトリーを指定する。
DPI が有効になったとき、本コマンドが設定されており、かつ本コマンドで指定されているディレクトリーにシグ
ネチャーが存在する場合には、ネットワーク経由のシグネチャーのダウンロードは行わず、外部メモリーまたは内
蔵不揮発性メモリー（ RTFS）のシグネチャーを使用する。
シグネチャーをダウンロードしたとき、 path に指定したディレクトリーが存在しなければ、ディレクトリーを自動
生成して、シグネチャーを保存する。  path に外部メモリーを指定している場合、指定したディレクトリーを含む外706 | コマンドリファレンス  | DPI

部メモリーが接続されていないとき、その間にダウンロードしたシグネチャーは保存できない。  その後外部メモリ
ーが接続された場合には、次のシグネチャー更新時にダウンロードしたシグネチャーが保存される。
次の更新を待たずにシグネチャーをダウンロードしたい場合には、 dpi signature download go コマンドを  force オプ
ションを付けて実行すると、シグネチャーの更新が無くとも直ちにダウンロード、および保存が可能となる。 path
で指定したディレクトリーに以前保存したシグネチャーが存在する場合には、新たにダウンロードしたシグネチャ
ーを既存のシグネチャーに上書きして保存する。
[ノート ]
path は半角 99文字以内。
シグネチャーは  "rt_dpi_( 機種名 ).ysig" というファイル名で保存される。例えば、 RTX1300 の場合には
"rt_dpi_rtx1300.ysig" となる。
RTX830 では  external-memory dpi signature directory  コマンドを使用する。
[設定例 ]
•内蔵不揮発性メモリー（ RTFS）の  dpi/signature ディレクトリーにシグネチャーを保存する
# dpi signature directory /dpi/signature
• microSD カードの  dpi/signature ディレクトリーにシグネチャーを保存する
# dpi signature directory sd1:/dpi/signature
[適用モデル ]
RTX1300, RTX840コマンドリファレンス  | DPI | 707

