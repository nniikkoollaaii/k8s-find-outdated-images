cat ./result.json

echo ""
echo ""
echo "Checking result ..."
echo ""
if [[ $(jq '. | length' result.json) != 1 ]]; then
  echo "Number of results not equals 1"; exit 1;
fi

if [[ $(jq '.["nginx:1.14.2"].Findings | length' result.json) != 3 ]]; then
  echo "Number of Findings for \"nginx:1.14.2\" not equals 3"; exit 1;
fi

if [[ "$(jq -r '.["nginx:1.14.2"].Findings[0].NotificationData.Email' result.json)" != "ns-admin@org.com" ]]; then
  echo "First finding for \"nginx:1.14.2\": Email value not expected"; exit 1;
fi
echo "Finished! All ok ..."