# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

docker run --rm  -it -e CF_Key="e6aefae8a9825d40dbf7eadb9a49967373847" -e CF_Email="mobiledgex.ops@mobiledgex.com"   -v "$(pwd)/out":/acme.sh    --net=host   neilpang/acme.sh  --issue -d vault.mobiledgex.net   --dns dns_cf
