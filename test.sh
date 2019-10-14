while [ 1 ]
do
  sleep 1
  kubectl get images -n demo-team | grep "Unknown" | wc -l
done
