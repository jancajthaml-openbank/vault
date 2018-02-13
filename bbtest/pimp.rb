require 'thread'

module Enumerable

  def par_each
    return unless block_given?

    backlog = Queue.new
    each { |work| backlog << work }

    (1..Enumerable.parallelism).map {
      backlog << nil
      Thread.new {
        while tick = backlog.deq
          yield tick
        end
      }
    }.each(&:join)
  end

  def par_map
    return unless block_given?

    result = []
    idx = 0
    backlog = Queue.new
    each { |work| backlog << work }

    (1..Enumerable.parallelism).map {
      backlog << nil
      Thread.new {
        while tick = backlog.deq
          i = idx
          idx += 1
          result[i] = (yield tick)
        end
      }
    }.each(&:join)

    result
  end

  def par_reject
    return unless block_given?

    result = []
    idx = 0
    backlog = Queue.new
    each { |work| backlog << work }

    (1..Enumerable.parallelism).map {
      backlog << nil
      Thread.new {
        while tick = backlog.deq
          i = idx
          idx += 1
          next if yield tick
          result[i] = tick
        end
      }
    }.each(&:join)

    result.compact
  end

  def par_select
    return unless block_given?

    result = []
    idx = 0
    backlog = Queue.new
    each { |work| backlog << work }

    (1..Enumerable.parallelism).map {
      backlog << nil
      Thread.new {
        while tick = backlog.deq
          i = idx
          idx += 1
          next unless yield tick
          result[i] = tick
        end
      }
    }.each(&:join)

    result.compact
  end

  private

  class << self; attr_accessor :parallelism ; end

  self.parallelism = (Integer(%x(getconf _NPROCESSORS_ONLN)) rescue 1) << 3

end
